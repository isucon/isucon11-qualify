<?php

declare(strict_types=1);

use Fig\Http\Message\StatusCodeInterface;
use Firebase\JWT\JWT;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Log\LoggerInterface;
use Slim\App;
use SlimSession\Helper as SessionHelper;

final class InitializeRequest
{
    public function __construct(public string $jiaServiceUrl)
    {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): self
    {
        try {
            /** @var array{jia_service_url: string} $data */
            $data = json_decode($json, true, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        return new self($data['jia_service_url']);
    }
}

final class InitializeResponse implements JsonSerializable
{
    public function __construct(public string $language)
    {
    }

    /**
     * @return array{language: string}
     */
    public function jsonSerialize(): array
    {
        return ['language' => $this->language];
    }
}

final class GetMeResponse implements JsonSerializable
{
    public function __construct(public string $jiaUserId)
    {
    }

    /**
     * @return array{jia_user_id: string}
     */
    public function jsonSerialize(): array
    {
        return ['jia_user_id' => $this->jiaUserId];
    }
}

return function (App $app) {
    $app->options('/{routes:.*}', function (Request $request, Response $response) {
        // CORS Pre-Flight OPTIONS Request Handler
        return $response;
    });

    $app->post('/initialize', Handler::class . ':postInitialize');

    $app->post('/api/auth', Handler::class . ':postAuthentication');
    $app->post('/api/signout', Handler::class . ':postSignout');
    $app->get('/api/user/me', Handler::class . ':getMe');
    $app->get('/api/isu', Handler::class . 'getIsuList');
    $app->post('/api/isu', Handler::class . 'postIsu');
    $app->get('/api/isu/{jia_isu_uuid}', Handler::class . 'getIsuId');
    $app->get('/api/isu/{jia_isu_uuid}/icon', Handler::class . 'getIsuIcon');
    $app->get('/api/isu/{jia_isu_uuid}/graph', Handler::class . 'getIsuGraph');
    $app->get('/api/condition/{jia_isu_uuid}', Handler::class . 'getIsuConditions');
    $app->get('/api/trend', Handler::class . 'getTrend');

    $app->post('/api/condition/{jia_isu_uuid}', Handler::class . 'postIsuCondition');

    $app->get('/', Handler::class . ':getIndex');
    $app->get('/isu/{jia_isu_uuid}', Handler::class . ':getIndex');
    $app->get('/isu/{jia_isu_uuid}/condition', Handler::class . ':getIndex');
    $app->get('/isu/{jia_isu_uuid}/graph', Handler::class . ':getIndex');
    $app->get('/register', Handler::class . ':getIndex');
    $app->get('/assets/{filename}', Handler::class . ':getAssets');
};

final class Handler
{
    private const CONDITION_LIMIT = 20;
    private const FRONTEND_CONTENTS_PATH = __DIR__ . '/../../public';
    private const JIA_JWT_SIGNING_KEY_PATH = __DIR__ . '/../../ec256-public.pem';
    private const DEFAULT_ICON_FILE_PATH = "../NoImage.jpg";
    private const DEFAULT_JIA_SERVICE_URL = "http://localhost:5000";
    private const MYSQL_ERR_NUM_DUPLICATE_ENTRY = 1062;
    private const CONDITION_LEVEL_INFO = "info";
    private const CONDITION_LEVEL_WARNING = "warning";
    private const CONDITION_LEVEL_CRITICAL = "critical";
    private const SCORE_CONDITION_LEVEL_INFO = 3;
    private const SCORE_CONDITION_LEVEL_WARNING = 2;
    private const SCORE_CONDITION_LEVEL_CRITICAL = 1;

    public function __construct(
        private PDO $dbh,
        private SessionHelper $session,
        private LoggerInterface $logger,
    ) {
    }

    /**
     * @return array{0: string, 1: int, 2: string}
     */
    private function getUserIdFromSession(): array
    {
        $jiaUserId = $this->session->get('jia_user_id');
        if (empty($jiaUserId)) {
            return ['', StatusCodeInterface::STATUS_UNAUTHORIZED, 'no session'];
        }

        $stmt = $this->dbh->prepare('SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?');
        if (!$stmt->execute([$jiaUserId])) {
            return ['', StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR, 'db error: ' . $this->dbh->errorInfo()[2]];
        }

        if ($stmt->fetch()[0] == 0) {
            return ['', StatusCodeInterface::STATUS_UNAUTHORIZED, 'not found: user'];
        }

        return [$jiaUserId, 0, ''];
    }

    private function getJiaServiceUrl(): string
    {
        throw new Exception('not implemented');
    }

    // POST /initialize
    // サービスを初期化
    public function postInitialize(Request $request, Response $response): Response
    {
        try {
            $initializeRequest = InitializeRequest::fromJson((string)$request->getBody());
        } catch (UnexpectedValueException) {
            $newResponse = $response->withStatus(StatusCodeInterface::STATUS_FORBIDDEN)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            $newResponse->getBody()->write('bad request body');

            return $newResponse;
        }

        $stderr = fopen('php://stderr', 'w');
        $process = proc_open(
            __DIR__ . '/../../sql/init.sh',
            [['pipe', 'r'], $stderr, $stderr],
            $pipes,
        );
        if ($process === false) {
            $this->logger->error('exec init.sh error: cannot open process');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $exitCode = proc_close($process);
        if ($exitCode !== 0) {
            $this->logger->error('exec init.sh error: exit with non-zero code');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $stmt = $this->dbh->prepare('INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)');
        if (!$stmt->execute(['jia_service_url', $initializeRequest->jiaServiceUrl])) {
            $this->logger->error('db error: ' . $this->dbh->errorInfo()[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        return $this->jsonResponse($response, new InitializeResponse(language: 'php'));
    }

    // POST /api/auth
    // サインアップ・サインイン
    public function postAuthentication(Request $request, Response $response): Response
    {
        $authorizationHeader = $request->getHeader('Authorization');
        if (count($authorizationHeader) < 1) {
            $newResponse = $response->withStatus(StatusCodeInterface::STATUS_FORBIDDEN)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            $newResponse->getBody()->write('forbidden');

            return $newResponse;
        }
        $reqJwt = mb_substr($authorizationHeader[0], mb_strlen('Bearer '));

        $jiaJwtSigningKey = file_get_contents(self::JIA_JWT_SIGNING_KEY_PATH);
        if ($jiaJwtSigningKey === false) {
            $this->get(LoggerInterface::class)->critical('failed to read file: ' . self::JIA_JWT_SIGNING_KEY_PATH);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        try {
            $token = JWT::decode($reqJwt, $jiaJwtSigningKey, ['ES256', 'ES384', 'ES512']);
        } catch (UnexpectedValueException) {
            $newResponse = $response->withStatus(StatusCodeInterface::STATUS_FORBIDDEN)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            $newResponse->getBody()->write('forbidden');

            return $newResponse;
        } catch (Exception $e) {
            $this->logger->error($e);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $jiaUserIdVar = $token->jia_user_id;
        if (empty($jiaUserIdVar)) {
            $newResponse = $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            $newResponse->getBody()->write('invalid JWT payload');

            return $newResponse;
        }
        $jiaUserId = (string)$jiaUserIdVar;

        $stmt = $this->dbh->prepare('INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)');
        if (!$stmt->execute([$jiaUserId])) {
            $this->logger->error('db error: ' . $this->dbh->errorInfo()[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $this->session->set('jia_user_id', $jiaUserId);

        return $response;
    }

    // POST /api/signout
    // サインアウト
    public function postSignout(Request $request, Response $response): Response
    {
        [$_, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse;
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $this->session->destroy();

        return $response;
    }

    // GET /api/user/me
    // サインインしている自分自身の情報を取得
    public function getMe(Request $request, Response $response): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse;
            }

            $this->logger->error($err);

            return $newResponse;
        }

        return $this->jsonResponse($response, new GetMeResponse(jiaUserId: $jiaUserId));
    }

    // GET /api/isu
    // ISUの一覧を取得
    public function getIsuList(Request $request, Response $response): Response
    {
        throw new Exception('not implemented');
    }

    // POST /api/isu
    // ISUを登録
    public function postIsu(Request $request, Response $response): Response
    {
        throw new Exception('not implemented');
    }

    // GET /api/isu/:jia_isu_uuid
    // ISUの情報を取得
    public function getIsuID(Request $request, Response $response, array $args): Response
    {
        throw new Exception('not implemented');
    }

    // GET /api/isu/:jia_isu_uuid/icon
    // ISUのアイコンを取得
    public function getIsuIcon(Request $request, Response $response, array $args): Response
    {
        throw new Exception('not implemented');
    }

    // GET /api/isu/:jia_isu_uuid/graph
    // ISUのコンディショングラフ描画のための情報を取得
    public function getIsuGraph(Request $request, Response $response, array $args): Response
    {
        throw new Exception('not implemented');
    }

    // グラフのデータ点を一日分生成
    private function generateIsuGraphResponse()
    {
        throw new Exception('not implemented');
    }

    // 複数のISUのコンディションからグラフの一つのデータ点を計算
    private function calculateGraphDataPoint()
    {
        throw new Exception('not implemented');
    }

    // GET /api/condition/:jia_isu_uuid
    // ISUのコンディションを取得
    public function getIsuConditions(Request $request, Response $response, array $args): Response
    {
        throw new Exception('not implemented');
    }

    // ISUのコンディションをDBから取得
    private function getIsuConditionsFromDB()
    {
        throw new Exception('not implemented');
    }

    // ISUのコンディションの文字列からコンディションレベルを計算
    private function calculateConditionLevel()
    {
        throw new Exception('not implemented');
    }

    // GET /api/trend
    // ISUの性格毎の最新のコンディション情報
    public function getTrend(Request $request, Response $response): Response
    {
        throw new Exception('not implemented');
    }

    // POST /api/condition/:jia_isu_uuid
    // ISUからのコンディションを受け取る
    public function postIsuCondition(Request $request, Response $response, array $args): Response
    {
        throw new Exception('not implemented');
    }

    // ISUのコンディションの文字列がcsv形式になっているか検証
    private function isValidConditionFormat()
    {
        throw new Exception('not implemented');
    }

    public function getIndex(Request $request, Response $response): Response
    {
        $response->getBody()->write(file_get_contents(self::FRONTEND_CONTENTS_PATH . '/index.html'));

        return $response;
    }

    public function getAssets(Request $request, Response $response, array $args): Response
    {
        $filePath = self::FRONTEND_CONTENTS_PATH . '/assets/' . $args['filename'];

        if (!file_exists($filePath)) {
            return $response->withStatus(404, 'File Not Found');
        }

        $mimeType = match (pathinfo($filePath, PATHINFO_EXTENSION)) {
            'js' => 'text/javascript',
            'css' => 'text/css',
            'svg' => 'image/svg+xml',
            default => 'text/html',
        };

        $newResponse = $response->withHeader('Content-Type', $mimeType . '; charset=UTF-8');
        $newResponse->getBody()->write(file_get_contents($filePath));

        return $newResponse;
    }

    private function jsonResponse(Response $response, JsonSerializable $data): Response
    {
        $responseBody = json_encode($data);
        if ($responseBody === false) {
            $this->logger->critical('failed to json_encode');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $response->getBody()->write($responseBody);

        return $response->withHeader('Content-Type', 'application/json');
    }
}
