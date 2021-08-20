require 'jwt'
require 'net/http'
require 'openssl'
require 'sinatra/base'
require 'uri'
require 'mysql2'
require 'mysql2-cs-bind'

module Isucondition
  class App < Sinatra::Base
    configure :development do
      require 'sinatra/reloader'
      register Sinatra::Reloader
    end

    SESSION_NAME = 'isucondition_ruby'
    CONDITION_LIMIT = 20
    FRONTEND_CONTENTS_PATH = '../public'
    JIA_JWT_SIGNING_KEY_PATH = '../ec256-public.pem'
    DEFAULT_ICON_FILE_PATH = '../NoImage.jpg'
    DEFAULT_JIA_SERVICE_URL = 'http://localhost:5000'

    MYSQL_ERR_NUM_DUPLICATE_ENTRY = 1062
    CONDITION_LEVEL_INFO = 'info'
    CONDITION_LEVEL_WARNING = 'warning'
    CONDITION_LEVEL_CRITICAL = 'critical'

    SCORE_CONDITION_LEVEL_INFO = 3
    SCORE_CONDITION_LEVEL_WARNING = 2
    SCORE_CONDITION_LEVEL_CRITICAL = 1

    set :session_secret, 'isucondition'
    set :sessions, key: SESSION_NAME

    set :public_folder, FRONTEND_CONTENTS_PATH
    set :protection, false  # IPアドレスでHTTPS接続した場合に一部機能が動かなくなるため無効化

    POST_ISU_CONDITION_TARGET_BASE_URL = ENV.fetch('POST_ISUCONDITION_TARGET_BASE_URL')
    JIA_JWT_SIGNING_KEY = OpenSSL::PKey::EC.new(File.read(JIA_JWT_SIGNING_KEY_PATH), '')

    class MySQLConnectionEnv
      def initialize
        @host = get_env('MYSQL_HOST', '127.0.0.1')
        @port = get_env('MYSQL_PORT', '3306')
        @user = get_env('MYSQL_USER', 'isucon')
        @db_name = get_env('MYSQL_DBNAME', 'isucondition')
        @password = get_env('MYSQL_PASS', 'isucon')
      end

      def connect_db
        Mysql2::Client.new(
          host: @host,
          port: @port,
          username: @user,
          database: @db_name,
          password: @password,
          charset: 'utf8mb4',
          database_timezone: :local,
          cast_booleans: true,
          symbolize_keys: true,
          reconnect: true,
        )
      end

      private

      def get_env(key, default)
        val = ENV.fetch(key, '')
        return val unless val.empty?
        default
      end
    end

    helpers do
      def json_params
        @json_params ||= JSON.parse(request.body.tap(&:rewind).read, symbolize_names: true)
      end

      def db
        Thread.current[:db] ||= MySQLConnectionEnv.new.connect_db
      end

      def db_transaction(&block)
        db.query('BEGIN')
        done = false
        retval = block.call
        db.query('COMMIT')
        done = true
        return retval
      ensure
        db.query('ROLLBACK') unless done
      end

      def halt_error(*args)
        content_type 'text/plain'
        halt(*args)
      end


      def user_id_from_session
        jia_user_id = session[:jia_user_id]
        return nil if !jia_user_id || jia_user_id.empty?
        count = db.xquery('SELECT COUNT(*) AS `cnt` FROM `user` WHERE `jia_user_id` = ?', jia_user_id).first
        return nil if count.fetch(:cnt).to_i.zero?

        jia_user_id
      end

      def jia_service_url
        config = db.xquery('SELECT * FROM `isu_association_config` WHERE `name` = ?', 'jia_service_url').first
        return DEFAULT_JIA_SERVICE_URL unless config
        config[:url]
      end

      # ISUのコンディションの文字列からコンディションレベルを計算
      def calculate_condition_level(condition)
        idx = -1
        warn_count = 0
        while idx
          idx = condition.index('=true', idx+1)
          warn_count += 1 if idx
        end

        case warn_count
        when 0
          CONDITION_LEVEL_INFO
        when 1, 2
          CONDITION_LEVEL_WARNING
        when 3
          CONDITION_LEVEL_CRITICAL
        else
          raise "unexpected warn count"
        end
      end

      # ISUのコンディションの文字列がcsv形式になっているか検証
      def valid_condition_format?(condition_str)
        keys = %w(is_dirty= is_overweight= is_broken=)
        value_true = 'true'
        value_false = 'false'

        idx_cond_str = 0
        keys.each_with_index do |key, idx_keys|
          return false unless condition_str[idx_cond_str..-1].start_with?(key)
          idx_cond_str += key.size
          case
          when condition_str[idx_cond_str..-1].start_with?(value_true)
            idx_cond_str += value_true.size
          when condition_str[idx_cond_str..-1].start_with?(value_false)
            idx_cond_str += value_false.size
          else
            return false
          end

          if idx_keys < (keys.size-1)
            return false unless condition_str[idx_cond_str] == ?,
            idx_cond_str += 1
          end
        end

        idx_cond_str == condition_str.size
      end
    end

    # サービスを初期化
    post '/initialize' do
      jia_service_url = begin
        json_params[:jia_service_url]
      rescue JSON::ParserError
        halt_error 400, 'bad request body'
      end
      halt_error 400, 'bad request body' unless jia_service_url

      system('../sql/init.sh', out: :err, exception: true)
      db.xquery(
        'INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)',
        'jia_service_url',
        jia_service_url,
      )

      content_type :json
      { language: 'ruby' }.to_json
    end

    # サインアップ・サインイン
    post '/api/auth' do
      req_jwt = request.env['HTTP_AUTHORIZATION']&.delete_prefix('Bearer ')
      token, _headers = begin
        JWT.decode(req_jwt, JIA_JWT_SIGNING_KEY, true, algorithm: 'ES256')
      rescue JWT::DecodeError
        halt_error 403, 'forbidden'
      end

      jia_user_id = token['jia_user_id']
      halt_error 400, 'invalid JWT payload' if !jia_user_id || !jia_user_id.is_a?(String)

      db.xquery('INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)', jia_user_id)

      session[:jia_user_id] = jia_user_id

      ''
    end

    # サインアウト
    post '/api/signout' do
      halt_error 401, 'you are not signed in' unless user_id_from_session
      session.destroy

      status 200
      ''
    end

    # サインインしている自分自身の情報を取得
    get '/api/user/me' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      content_type :json
      { jia_user_id: jia_user_id }.to_json
    end

    # ISUの一覧を取得
    get '/api/isu' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      response_list = db_transaction do
        isu_list = db.xquery('SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC', jia_user_id)
        isu_list.map do |isu|
          last_condition = db.xquery('SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1', isu.fetch(:jia_isu_uuid)).first

          formatted_condition = last_condition ? {
            jia_isu_uuid: last_condition.fetch(:jia_isu_uuid),
            isu_name: isu.fetch(:name),
            timestamp: last_condition.fetch(:timestamp).to_i,
            is_sitting: last_condition.fetch(:is_sitting),
            condition: last_condition.fetch(:condition),
            condition_level: calculate_condition_level(last_condition.fetch(:condition)),
            message: last_condition.fetch(:message),
          } : nil

          {
            id: isu.fetch(:id),
            jia_isu_uuid: isu.fetch(:jia_isu_uuid),
            name: isu.fetch(:name),
            character: isu.fetch(:character),
            latest_isu_condition: formatted_condition,
          }
        end
      end

      content_type :json
      response_list.to_json
    end

    # ISUを登録
    post '/api/isu' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      jia_isu_uuid = params[:jia_isu_uuid]
      isu_name = params[:isu_name]

      fh = params[:image]
      halt_error 400, 'bad format: icon' if fh && (!fh.kind_of?(Hash) || !fh[:tempfile].is_a?(Tempfile))

      use_default_image = fh.nil?
      image = use_default_image ? File.binread(DEFAULT_ICON_FILE_PATH) : fh.fetch(:tempfile).binmode.read

      isu = db_transaction do
        begin
          db.xquery(
            "INSERT INTO `isu` (`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)".b,
            jia_isu_uuid.b, isu_name.b, image, jia_user_id.b,
          )
        rescue Mysql2::Error => e
          if e.error_number == MYSQL_ERR_NUM_DUPLICATE_ENTRY
            halt_error 409, "duplicated: isu"
          end

          raise
        end

        target_url = URI.parse("#{jia_service_url}/api/activate")
        http = Net::HTTP.new(target_url.host, target_url.port)
        http.use_ssl = target_url.scheme == 'https'
        res = http.start do
          req = Net::HTTP::Post.new(target_url.path)
          req['content-type'] = 'application/json'
          req.body = { target_base_url: POST_ISU_CONDITION_TARGET_BASE_URL, isu_uuid: jia_isu_uuid }.to_json
          http.request(req)
        end
        if res.code != '202'
          request.env['rack.logger'].warn "JIAService returned error: status code #{res.code}, message #{res.body.inspect}"
          halt_error res.code.to_i, 'JIAService returned error'
        end
        isu_from_jia = JSON.parse(res.body, symbolize_names: true)

        db.xquery('UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?', isu_from_jia.fetch(:character), jia_isu_uuid)
        db.xquery('SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?', jia_user_id, jia_isu_uuid).first
      end

      status 201
      content_type :json
      {
        id: isu.fetch(:id),
        jia_isu_uuid: isu.fetch(:jia_isu_uuid),
        name: isu.fetch(:name),
        character: isu.fetch(:character),
        jia_user_id: isu.fetch(:jia_user_id),
      }.to_json
    end

    # ISUの情報を取得
    get '/api/isu/:jia_isu_uuid' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      jia_isu_uuid = params[:jia_isu_uuid]
      isu = db.xquery('SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?', jia_user_id, jia_isu_uuid).first
      halt_error 404, 'not found: isu' unless isu

      content_type :json
      {
        id: isu.fetch(:id),
        jia_isu_uuid: isu.fetch(:jia_isu_uuid),
        name: isu.fetch(:name),
        character: isu.fetch(:character),
        jia_user_id: isu.fetch(:jia_user_id),
      }.to_json
    end

    # ISUのアイコンを取得
    get '/api/isu/:jia_isu_uuid/icon' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      jia_isu_uuid = params[:jia_isu_uuid]
      isu = db.xquery('SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?', jia_user_id, jia_isu_uuid).first
      halt_error 404, 'not found: isu' unless isu

      isu.fetch(:image)
    end

    # ISUのコンディショングラフ描画のための情報を取得
    get '/api/isu/:jia_isu_uuid/graph' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      jia_isu_uuid = params[:jia_isu_uuid]
      datetime_str = params[:datetime]
      halt_error 400, 'missing: datetime' if !datetime_str || datetime_str.empty?
      datetime = Time.at(Integer(datetime_str)) rescue halt_error(400, 'bad format: datetime')
      date = Time.new(datetime.year, datetime.month, datetime.day, datetime.hour, 0, 0)


      res = db_transaction do
        cnt = db.xquery('SELECT COUNT(*) AS `cnt` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?', jia_user_id, jia_isu_uuid).first
        halt_error 404, 'not found: isu' if cnt.fetch(:cnt) == 0

        generate_isu_graph_response(jia_isu_uuid, date)
      end

      content_type :json
      res.to_json
    end

    # グラフのデータ点を一日分生成
    def generate_isu_graph_response(jia_isu_uuid, graph_date)
      rows = db.xquery('SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC', jia_isu_uuid)

      data_points = []
      start_time_in_this_hour = Time.at(0)
      conditions_in_this_hour = []
      timestamps_in_this_hour = []

      rows.each do |condition|
        timestamp = condition.fetch(:timestamp)
        truncated_condition_time = Time.new(timestamp.year, timestamp.month, timestamp.day, timestamp.hour, 0, 0)
        if truncated_condition_time != start_time_in_this_hour
          unless conditions_in_this_hour.empty?
            data = calculate_graph_data_point(conditions_in_this_hour)
            data_points.push(
              jia_isu_uuid: jia_isu_uuid,
              start_at: start_time_in_this_hour,
              data: data,
              condition_timestamps: timestamps_in_this_hour,
            )
          end

          start_time_in_this_hour = truncated_condition_time
          conditions_in_this_hour = []
          timestamps_in_this_hour = []
        end
        conditions_in_this_hour.push(condition)
        timestamps_in_this_hour.push(condition.fetch(:timestamp).to_i)
      end

      unless conditions_in_this_hour.empty?
        data = calculate_graph_data_point(conditions_in_this_hour)
        data_points.push(
          jia_isu_uuid: jia_isu_uuid,
          start_at: start_time_in_this_hour,
          data: data,
          condition_timestamps: timestamps_in_this_hour,
        )
      end

      end_time = graph_date + (3600 * 24)
      start_index = data_points.size
      end_next_index = data_points.size

      data_points.each_with_index do |graph, i|
        start_index = i if start_index == data_points.size && graph.fetch(:start_at) >= graph_date
        end_next_index = i if end_next_index == data_points.size && graph.fetch(:start_at) > end_time
      end

      filtered_data_points = []
      filtered_data_points = data_points[start_index...end_next_index] if start_index < end_next_index

      response_list = []
      index = 0
      this_time = graph_date

      while this_time < (graph_date + (3600*24))
        data = nil
        timestamps = []
        if index < filtered_data_points.size
          data_with_info = filtered_data_points[index]
          if data_with_info.fetch(:start_at) == this_time
            data = data_with_info.fetch(:data)
            timestamps = data_with_info.fetch(:condition_timestamps)
            index += 1
          end
        end

        response_list.push(
          start_at: this_time.to_i,
          end_at: (this_time + 3600).to_i,
          data: data,
          condition_timestamps: timestamps,
        )
        this_time += 3600
      end

      response_list
    end

    # 複数のISUのコンディションからグラフの一つのデータ点を計算
    def calculate_graph_data_point(isu_conditions)
      conditions_count = {
        'is_broken' => 0,
        'is_dirty' => 0,
        'is_overweight' => 0,
      }
      raw_score = 0

      isu_conditions.each do |condition|
        bad_conditions_count = 0

        unless valid_condition_format?(condition.fetch(:condition))
          raise "invalid condition format"
        end

        condition.fetch(:condition).split(',').each do |cond_str|
          condition_name, value = cond_str.split('=')
          if value == 'true'
            conditions_count[condition_name] += 1
            bad_conditions_count += 1
          end
        end

        case
        when bad_conditions_count >= 3
          raw_score += SCORE_CONDITION_LEVEL_CRITICAL
        when bad_conditions_count >= 1
          raw_score += SCORE_CONDITION_LEVEL_WARNING
        else
          raw_score += SCORE_CONDITION_LEVEL_INFO
        end
      end

      sitting_count = 0
      isu_conditions.each do |condition|
        sitting_count += 1 if condition.fetch(:is_sitting)
      end

      isu_conditions_length = isu_conditions.size
      score = raw_score * 100 / 3 / isu_conditions_length
      sitting_percentage = sitting_count * 100 / isu_conditions_length
      is_broken_percentage = conditions_count.fetch('is_broken') * 100 / isu_conditions_length
      is_overweight_percentage = conditions_count.fetch('is_overweight') * 100 / isu_conditions_length
      is_dirty_percentage = conditions_count.fetch('is_dirty') * 100 / isu_conditions_length

      {
        score: score,
        percentage: {
          sitting: sitting_percentage,
          is_broken: is_broken_percentage,
          is_overweight: is_overweight_percentage,
          is_dirty: is_dirty_percentage,
        },
      }
    end

    # ISUのコンディションを取得
    get '/api/condition/:jia_isu_uuid' do
      jia_user_id = user_id_from_session
      halt_error 401, 'you are not signed in' unless jia_user_id

      jia_isu_uuid = params[:jia_isu_uuid]
      halt_error 400, 'missing: jia_isu_uuid' if !jia_isu_uuid || jia_isu_uuid.empty?

      end_time_integer = params[:end_time].yield_self { |_| Integer(_) } rescue halt_error(400, 'bad format: end_time')
      end_time = Time.at(end_time_integer)

      condition_level_csv = params[:condition_level]
      halt_error 400, 'missing: condition_level' if !condition_level_csv || condition_level_csv.empty?
      condition_level = Set.new(condition_level_csv.split(','))

      start_time_str = params[:start_time]
      start_time = Time.at(start_time_str && !start_time_str.empty? ? Integer(start_time_str) : 0) rescue halt_error(400, 'bad format: start_time')

      isu = db.xquery('SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?', jia_isu_uuid, jia_user_id).first
      halt_error 404, 'not found: isu' unless isu
      isu_name = isu.fetch(:name)

      conditions_response = get_isu_conditions_from_db(
        jia_isu_uuid,
        end_time,
        condition_level,
        start_time,
        CONDITION_LIMIT,
        isu_name,
      )

      content_type :json
      conditions_response.to_json
    end

    # ISUのコンディションをDBから取得
    def get_isu_conditions_from_db(jia_isu_uuid, end_time, condition_level, start_time, limit, isu_name)
      conditions = if start_time.to_i == 0
        db.xquery(
          'SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? AND `timestamp` < ? ORDER BY `timestamp` DESC',
          jia_isu_uuid,
          end_time,
        )
      else
        db.xquery(
          'SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? AND `timestamp` < ? AND ? <= `timestamp` ORDER BY `timestamp` DESC',
          jia_isu_uuid,
          end_time,
          start_time,
        )
      end

      conditions_response = conditions.map do |c|
        c_level = calculate_condition_level(c.fetch(:condition))
        if condition_level.include?(c_level)
          {
            jia_isu_uuid: c.fetch(:jia_isu_uuid),
            isu_name: isu_name,
            timestamp: c.fetch(:timestamp).to_i,
            is_sitting: c.fetch(:is_sitting),
            condition: c.fetch(:condition),
            condition_level: c_level,
            message: c.fetch(:message),
          }
        else
          nil
        end
      end.compact

      conditions_response = conditions_response[0, limit] if conditions_response.size > limit
      conditions_response
    end

    # ISUの性格毎の最新のコンディション情報
    get '/api/trend' do
      character_list = db.query('SELECT `character` FROM `isu` GROUP BY `character`')

      res = character_list.map do |character|
        isu_list = db.xquery('SELECT * FROM `isu` WHERE `character` = ?', character.fetch(:character))
        character_info_isu_conditions = []
        character_warning_isu_conditions = []
        character_critical_isu_conditions = []

        isu_list.each do |isu|
          conditions = db.xquery('SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY timestamp DESC', isu.fetch(:jia_isu_uuid)).to_a
          unless conditions.empty?
            isu_last_condition = conditions.first
            condition_level = calculate_condition_level(isu_last_condition.fetch(:condition))
            trend_condition = { isu_id: isu.fetch(:id), timestamp: isu_last_condition.fetch(:timestamp).to_i }
            case condition_level
            when 'info'
              character_info_isu_conditions.push(trend_condition)
            when 'warning'
              character_warning_isu_conditions.push(trend_condition)
            when 'critical'
              character_critical_isu_conditions.push(trend_condition)
            end
          end
        end

        character_info_isu_conditions.sort! { |a,b| b.fetch(:timestamp) <=> a.fetch(:timestamp) }
        character_warning_isu_conditions.sort! { |a,b| b.fetch(:timestamp) <=> a.fetch(:timestamp) }
        character_critical_isu_conditions.sort! { |a,b| b.fetch(:timestamp) <=> a.fetch(:timestamp) }

        {
          character: character.fetch(:character),
          info: character_info_isu_conditions,
          warning: character_warning_isu_conditions,
          critical: character_critical_isu_conditions,
        }
      end

      content_type :json
      res.to_json
    end

    # ISUからのコンディションを受け取る
    post '/api/condition/:jia_isu_uuid' do
      # TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
      drop_probability = 0.9
      if rand <= drop_probability
        request.env['rack.logger'].warn 'drop post isu condition request'
        halt_error 202, ''
      end

      jia_isu_uuid = params[:jia_isu_uuid]
      halt_error 400, 'missing: jia_isu_uuid' if !jia_isu_uuid || jia_isu_uuid.empty?

      begin
        json_params
      rescue JSON::ParserError
        halt_error 400, 'bad request body'
      end
      halt_error 400, 'bad request body' unless json_params.kind_of?(Array)
      halt_error 400, 'bad request body' if json_params.empty?

      db_transaction do
        count = db.xquery('SELECT COUNT(*) AS `cnt` FROM `isu` WHERE `jia_isu_uuid` = ?', jia_isu_uuid).first
        halt_error 404, 'not found: isu' if count.fetch(:cnt).zero?

        json_params.each do |cond|
          timestamp = Time.at(cond.fetch(:timestamp))
          halt_error 400, 'bad request body' unless valid_condition_format?(cond.fetch(:condition))

          db.xquery(
            'INSERT INTO `isu_condition` (`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`) VALUES (?, ?, ?, ?, ?)',
            jia_isu_uuid,
            timestamp,
            cond.fetch(:is_sitting),
            cond.fetch(:condition),
            cond.fetch(:message),
          )
        end
      end

      status 202
      ''
    end

    %w(
      /
      /register
      /isu/:jia_isu_uuid
      /isu/:jia_isu_uuid/condition
      /isu/:jia_isu_uuid/graph
    ).each do |_|
      get _ do
        content_type :html
        File.read(File.join(FRONTEND_CONTENTS_PATH, 'index.html'))
      end
    end
  end
end
