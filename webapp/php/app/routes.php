<?php

declare(strict_types=1);

use Fig\Http\Message\StatusCodeInterface;
use Firebase\JWT\JWT;
use GuzzleHttp\ClientInterface as HttpClient;
use Psr\Http\Client\ClientExceptionInterface as HttpClientException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Http\Message\UploadedFileInterface;
use Psr\Log\LoggerInterface;
use Slim\App;
use SlimSession\Helper as SessionHelper;

final class Isu implements JsonSerializable
{
    public function __construct(
        public ?int $id,
        public ?string $jiaIsuUuid,
        public ?string $name,
        public ?string $image,
        public ?string $character,
        public ?string $jiaUserId,
        public ?DateTimeInterface $createdAt,
        public ?DateTimeInterface $updatedAt,
    ) {
    }

    /**
     * @param array{id?: string, jia_isu_uuid?: string, name?: string, image?: string, character?: string, jia_user_id?: string, created_at?: string, updated_at?: string} $dbRow
     * @throws Exception
     */
    public static function fromDbRow(array $dbRow): self
    {
        return new self(
            isset($dbRow['id']) ? (int)$dbRow['id'] : null,
            $dbRow['jia_isu_uuid'] ?? null,
            $dbRow['name'] ?? null,
            $dbRow['image'] ?? null,
            $dbRow['character'] ?? null,
            $dbRow['jia_user_id'] ?? null,
            isset($dbRow['created_at']) ? new DateTimeImmutable($dbRow['created_at']) : null,
            isset($dbRow['updated_at']) ? new DateTimeImmutable($dbRow['updated_at']) : null,
        );
    }

    /**
     * @return array{id: ?int, jia_isu_uuid: ?string, name: ?string, character: ?string}
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'jia_isu_uuid' => $this->jiaIsuUuid,
            'name' => $this->name,
            'character' => $this->character,
        ];
    }
}

final class IsuFromJia
{
    public function __construct(
        public ?string $character,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): self
    {
        try {
            $data = json_decode($json, true, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        if (!isset($data['character'])) {
            throw new UnexpectedValueException();
        }

        return new self($data['character'] ?? null);
    }
}

final class GetIsuListResponse implements JsonSerializable
{
    public function __construct(
        public int $id,
        public string $jiaIsuUuid,
        public string $name,
        public string $character,
        public ?GetIsuConditionResponse $latestIsuCondition,
    ) {
    }

    /**
     * @return array{id: int, jia_isu_uuid: string, name: string, character: string, latest_isu_condition: ?GetIsuConditionResponse}
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'jia_isu_uuid' => $this->jiaIsuUuid,
            'name' => $this->name,
            'character' => $this->character,
            'latest_isu_condition' => $this->latestIsuCondition,
        ];
    }
}

final class IsuCondition
{
    public function __construct(
        public ?int $id,
        public ?string $jiaIsuUuid,
        public ?DateTimeInterface $timestamp,
        public ?bool $isSitting,
        public ?string $condition,
        public ?string $message,
        public ?DateTimeInterface $createdAt,
    ) {
    }

    /**
     * @param array{id?: string, jia_isu_uuid?: string, timestamp?: string, is_sitting?: string, condition?: string, message?: string, created_at?: string} $dbRow
     * @throws Exception
     */
    public static function fromDbRow(array $dbRow): self
    {
        return new self(
            isset($dbRow['id']) ? (int)$dbRow['id'] : null,
            $dbRow['jia_isu_uuid'] ?? null,
            isset($dbRow['timestamp']) ? new DateTimeImmutable($dbRow['timestamp']) : null,
            isset($dbRow['is_sitting']) ? (bool)$dbRow['is_sitting'] : null,
            $dbRow['condition'] ?? null,
            $dbRow['message'] ?? null,
            isset($dbRow['created_at']) ? new DateTimeImmutable($dbRow['created_at']) : null,
        );
    }
}

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
            $data = json_decode($json, true, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        if (!isset($data['jia_service_url'])) {
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

final class GraphResponse implements JsonSerializable
{
    /**
     * @param array<int> $conditionTimestamps
     */
    public function __construct(
        public int $startAt,
        public int $endAt,
        public ?GraphDataPoint $data,
        public array $conditionTimestamps,
    ) {
    }

    /**
     * @return array{start_at: int, end_at: int, data: ?GraphDataPoint, condition_timestamps: array<int>}
     */
    public function jsonSerialize(): array
    {
        return [
            'start_at' => $this->startAt,
            'end_at' => $this->endAt,
            'data' => $this->data,
            'condition_timestamps' => $this->conditionTimestamps,
        ];
    }
}

final class GraphDataPoint implements JsonSerializable
{
    public function __construct(
        public int $score,
        public ConditionsPercentage $percentage,
    ) {
    }

    /**
     * @return array{score: int, percentage: ConditionsPercentage}
     */
    public function jsonSerialize(): array
    {
        return [
            'score' => $this->score,
            'percentage' => $this->percentage,
        ];
    }
}

final class ConditionsPercentage implements JsonSerializable
{
    public function __construct(
        public int $sitting,
        public int $isBroken,
        public int $isDirty,
        public int $isOverweight,
    ) {
    }

    /**
     * @return array{sitting: int, is_broken: int, is_dirty: int, is_overweight: int}
     */
    public function jsonSerialize(): array
    {
        return [
            'sitting' => $this->sitting,
            'is_broken' => $this->isBroken,
            'is_dirty' => $this->isDirty,
            'is_overweight' => $this->isOverweight,
        ];
    }
}

final class GraphDataPointWithInfo
{
    /**
     * @param array<int> $conditionTimestamps
     */
    public function __construct(
        public string $jiaIsuUuid,
        public DateTimeInterface $startAt,
        public GraphDataPoint $data,
        public array $conditionTimestamps,
    ) {
    }
}

final class GetIsuConditionResponse implements JsonSerializable
{
    public function __construct(
        public string $jiaIsuUuid,
        public string $isuName,
        public int $timestamp,
        public bool $isSitting,
        public string $condition,
        public string $conditionLevel,
        public string $message,
    ) {
    }

    /**
     * @return array{jia_isu_uuid: string, isu_name: string, timestamp: int, is_sitting: bool, condition: string, condition_level: string, message: string}
     */
    public function jsonSerialize(): array
    {
        return [
            'jia_isu_uuid' => $this->jiaIsuUuid,
            'isu_name' => $this->isuName,
            'timestamp' => $this->timestamp,
            'is_sitting' => $this->isSitting,
            'condition' => $this->condition,
            'condition_level' => $this->conditionLevel,
            'message' => $this->message,
        ];
    }
}

final class TrendResponse implements JsonSerializable
{
    /**
     * @param array<TrendCondition> $info
     * @param array<TrendCondition> $warning
     * @param array<TrendCondition> $critical
     */
    public function __construct(
        public string $character,
        public array $info,
        public array $warning,
        public array $critical,
    ) {
    }

    /**
     * @return array{character: string, info: array<TrendCondition>, warning: array<TrendCondition>, critical: array<TrendCondition>}
     */
    public function jsonSerialize(): array
    {
        return [
            'character' => $this->character,
            'info' => $this->info,
            'warning' => $this->warning,
            'critical' => $this->critical,
        ];
    }
}

final class TrendCondition implements JsonSerializable
{
    public function __construct(
        public int $id,
        public int $timestamp,
    ) {
    }

    /**
     * @return array{isu_id: int, timestamp: int}
     */
    public function jsonSerialize(): array
    {
        return [
            'isu_id' => $this->id,
            'timestamp' => $this->timestamp,
        ];
    }
}

final class PostIsuConditionRequest
{
    public function __construct(
        public bool $isSitting,
        public string $condition,
        public string $message,
        public int $timestamp,
    ) {
    }

    /**
     * @return array<self>
     * @throws UnexpectedValueException
     */
    public static function listFromJson(string $json): array
    {
        try {
            $data = json_decode($json, true, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        /** @var array<self> $list */
        $list = [];
        foreach ($data as $condition) {
            if (
                !isset($condition['is_sitting']) ||
                !isset($condition['condition']) ||
                !isset($condition['message']) ||
                !isset($condition['timestamp'])
            ) {
                throw new UnexpectedValueException();
            }

            $list[] = new self(
                $condition['is_sitting'],
                $condition['condition'],
                $condition['message'],
                $condition['timestamp']
            );
        }

        return $list;
    }
}

final class JiaServiceRequest implements JsonSerializable
{
    public function __construct(
        public string $targetBaseUrl,
        public string $isuUuid,
    ) {
    }

    /**
     * @return array{target_base_url: string, isu_uuid: string}
     */
    public function jsonSerialize(): array
    {
        return [
            'target_base_url' => $this->targetBaseUrl,
            'isu_uuid' => $this->isuUuid,
        ];
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
    $app->get('/api/isu', Handler::class . ':getIsuList');
    $app->post('/api/isu', Handler::class . ':postIsu');
    $app->get('/api/isu/{jia_isu_uuid}', Handler::class . ':getIsuId');
    $app->get('/api/isu/{jia_isu_uuid}/icon', Handler::class . ':getIsuIcon');
    $app->get('/api/isu/{jia_isu_uuid}/graph', Handler::class . ':getIsuGraph');
    $app->get('/api/condition/{jia_isu_uuid}', Handler::class . ':getIsuConditions');
    $app->get('/api/trend', Handler::class . ':getTrend');

    $app->post('/api/condition/{jia_isu_uuid}', Handler::class . ':postIsuCondition');

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
    private const DEFAULT_ICON_FILE_PATH = __DIR__ . '/../../NoImage.jpg';
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
        private HttpClient $httpClient,
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

        try {
            $stmt = $this->dbh->prepare('SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?');
            $stmt->execute([$jiaUserId]);
            $count = $stmt->fetch()[0];
        } catch (PDOException $e) {
            return ['', StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR, 'db error: ' . $e->errorInfo[2]];
        }

        if ($count == 0) {
            return ['', StatusCodeInterface::STATUS_UNAUTHORIZED, 'not found: user'];
        }

        return [$jiaUserId, 0, ''];
    }

    private function getJiaServiceUrl(): string
    {
        try {
            $stmt = $this->dbh->prepare('SELECT * FROM `isu_association_config` WHERE `name` = ?');
            $stmt->execute(['jia_service_url']);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->logger->warning($e->errorInfo[2]);

            return self::DEFAULT_JIA_SERVICE_URL;
        }

        if (count($rows) === 0) {
            return self::DEFAULT_JIA_SERVICE_URL;
        }

        return $rows[0]['url'];
    }

    /**
     * POST /initialize
     * サービスを初期化
     */
    public function postInitialize(Request $request, Response $response): Response
    {
        try {
            $initializeRequest = InitializeRequest::fromJson((string)$request->getBody());
        } catch (UnexpectedValueException) {
            $response->getBody()->write('bad request body');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
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

        try {
            $stmt = $this->dbh->prepare('INSERT INTO `isu_association_config` (`name`, `url`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `url` = VALUES(`url`)');
            $stmt->execute(['jia_service_url', $initializeRequest->jiaServiceUrl]);
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        return $this->jsonResponse($response, new InitializeResponse(language: 'php'));
    }

    /**
     * POST /api/auth
     * サインアップ・サインイン
     */
    public function postAuthentication(Request $request, Response $response): Response
    {
        $authorizationHeader = $request->getHeader('Authorization');
        if (count($authorizationHeader) < 1) {
            $response->getBody()->write('forbidden');

            return $response->withStatus(StatusCodeInterface::STATUS_FORBIDDEN)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $reqJwt = preg_replace('/\ABearer /', '', $authorizationHeader[0]);

        $jiaJwtSigningKey = file_get_contents(self::JIA_JWT_SIGNING_KEY_PATH);
        if ($jiaJwtSigningKey === false) {
            $this->logger->critical('failed to read file: ' . self::JIA_JWT_SIGNING_KEY_PATH);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        try {
            $claims = JWT::decode($reqJwt, $jiaJwtSigningKey, ['ES256', 'ES384', 'ES512']);
        } catch (UnexpectedValueException) {
            $response->getBody()->write('forbidden');

            return $response->withStatus(StatusCodeInterface::STATUS_FORBIDDEN)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        } catch (Exception $e) {
            $this->logger->error($e);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if (!property_exists($claims, 'jia_user_id')) {
            $response->getBody()->write('invalid JWT payload');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        $jiaUserId = $claims->jia_user_id;

        if (!is_string($jiaUserId)) {
            $response->getBody()->write('invalid JWT payload');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        try {
            $stmt = $this->dbh->prepare('INSERT IGNORE INTO user (`jia_user_id`) VALUES (?)');
            $stmt->execute([$jiaUserId]);
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $this->session->set('jia_user_id', $jiaUserId);

        return $response;
    }

    /**
     * POST /api/signout
     * サインアウト
     */
    public function postSignout(Request $request, Response $response): Response
    {
        [$_, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $this->session->destroy();

        return $response;
    }

    /**
     * GET /api/user/me
     * サインインしている自分自身の情報を取得
     */
    public function getMe(Request $request, Response $response): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        return $this->jsonResponse($response, new GetMeResponse(jiaUserId: $jiaUserId));
    }

    /**
     * GET /api/isu
     * ISUの一覧を取得
     */
    public function getIsuList(Request $request, Response $response): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $this->dbh->beginTransaction();

        try {
            $stmt = $this->dbh->prepare('SELECT * FROM `isu` WHERE `jia_user_id` = ? ORDER BY `id` DESC');
            $stmt->execute([$jiaUserId]);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        /** @var array<Isu> $isuList */
        $isuList = [];
        foreach ($rows as $row) {
            $isuList[] = Isu::fromDbRow($row);
        }

        /** @var array<GetIsuListResponse> $responseList */
        $responseList = [];
        foreach ($isuList as $isu) {
            try {
                $stmt = $this->dbh->prepare('SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` DESC LIMIT 1');
                $stmt->execute([$isu->jiaIsuUuid]);
                $rows = $stmt->fetchAll();
            } catch (PDOException $e) {
                $this->dbh->rollBack();
                $this->logger->error('db error: ' . $e->errorInfo[2]);

                return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
            }

            $formattedCondition = null;
            if (count($rows) != 0) {
                $lastCondition = IsuCondition::fromDbRow($rows[0]);
                try {
                    $conditionLevel = $this->calculateConditionLevel($lastCondition->condition);
                } catch (UnexpectedValueException $e) {
                    $this->dbh->rollBack();
                    $this->logger->error($e->getMessage());

                    return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
                }

                $formattedCondition = new GetIsuConditionResponse(
                    jiaIsuUuid: $lastCondition->jiaIsuUuid,
                    isuName: $isu->name,
                    timestamp: $lastCondition->timestamp->getTimestamp(),
                    isSitting: $lastCondition->isSitting,
                    condition: $lastCondition->condition,
                    conditionLevel: $conditionLevel,
                    message: $lastCondition->message,
                );
            }

            $res = new GetIsuListResponse(
                id: $isu->id,
                jiaIsuUuid: $isu->jiaIsuUuid,
                name: $isu->name,
                character: $isu->character,
                latestIsuCondition: $formattedCondition,
            );
            $responseList[] = $res;
        }

        $this->dbh->commit();

        return $this->jsonResponse($response, $responseList);
    }

    /**
     * POST /api/isu
     * ISUを登録
     */
    public function postIsu(Request $request, Response $response): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $useDefaultImage = false;

        $params = (array)$request->getParsedBody();
        $jiaIsuUuid = $params['jia_isu_uuid'];
        $isuName = $params['isu_name'];

        $uploadedFiles = $request->getUploadedFiles();
        if (isset($uploadedFiles['image'])) {
            /** @var UploadedFileInterface $imageFile */
            $imageFile = $uploadedFiles['image'];

            if ($imageFile->getError() !== UPLOAD_ERR_OK) {
                $response->getBody()->write('bad format: icon');

                return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                    ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }
        } else {
            $useDefaultImage = true;
        }

        if ($useDefaultImage) {
            $image = file_get_contents(self::DEFAULT_ICON_FILE_PATH);
            if ($image === false) {
                $this->logger->error('failed to read file: ' . self::JIA_JWT_SIGNING_KEY_PATH);

                return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
            }
        } else {
            try {
                $image = $imageFile->getStream()->getContents();
            } catch (RuntimeException $e) {
                $this->logger->error($e);

                return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
            }
        }

        $this->dbh->beginTransaction();

        try {
            $stmt = $this->dbh->prepare('INSERT INTO `isu`' .
                                        '	(`jia_isu_uuid`, `name`, `image`, `jia_user_id`) VALUES (?, ?, ?, ?)');
            $stmt->execute([$jiaIsuUuid, $isuName, $image, $jiaUserId]);
        } catch (PDOException $e) {
            $this->dbh->rollBack();

            if ($e->errorInfo[1] === self::MYSQL_ERR_NUM_DUPLICATE_ENTRY) {
                $response->getBody()->write('duplicated: isu');

                return $response->withStatus(StatusCodeInterface::STATUS_CONFLICT)
                    ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $targetUrl = $this->getJiaServiceUrl() . '/api/activate';
        $postIsuConditionTargetBaseUrl = getenv('POST_ISUCONDITION_TARGET_BASE_URL');
        if (!$postIsuConditionTargetBaseUrl) {
            $this->dbh->rollBack();
            $this->logger->critical('missing: POST_ISUCONDITION_TARGET_BASE_URL');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $body = new JiaServiceRequest($postIsuConditionTargetBaseUrl, $jiaIsuUuid);
        try {
            $res = $this->httpClient->request('POST', $targetUrl, ['json' => $body, 'http_errors' => false]);
        } catch (HttpClientException $e) {
            $this->dbh->rollBack();
            $this->logger->error('failed to request to JIAService: ' . $e->getMessage());

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $resBody = $res->getBody();
        $statusCode = $res->getStatusCode();
        if ($statusCode !== StatusCodeInterface::STATUS_ACCEPTED) {
            $this->dbh->rollBack();
            $this->logger->error(sprintf('JIAService returned error: status code %d, message: %s', $statusCode, (string)$resBody));

            $response->getBody()->write('JIAService returned error');

            return $response->withStatus($statusCode)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        try {
            $isuFromJia = IsuFromJia::fromJson((string)$resBody);
        } catch (UnexpectedValueException) {
            $this->dbh->rollBack();
            $this->logger->error('failed to json_encode');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        try {
            $stmt = $this->dbh->prepare('UPDATE `isu` SET `character` = ? WHERE  `jia_isu_uuid` = ?');
            $stmt->execute([$isuFromJia->character, $jiaIsuUuid]);

            $stmt = $this->dbh->prepare('SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?');
            $stmt->execute([$jiaUserId, $jiaIsuUuid]);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if (count($rows) === 0) {
            $this->dbh->rollBack();
            $this->logger->error('db error: failed to insert isu');

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $isu = Isu::fromDbRow($rows[0]);

        $this->dbh->commit();

        return $this->jsonResponse($response, $isu, StatusCodeInterface::STATUS_CREATED);
    }

    /**
     * GET /api/isu/:jia_isu_uuid
     * ISUの情報を取得
     */
    public function getIsuID(Request $request, Response $response, array $args): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $jiaIsuUuid = $args['jia_isu_uuid'];

        try {
            $stmt = $this->dbh->prepare('SELECT * FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?');
            $stmt->execute([$jiaUserId, $jiaIsuUuid]);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if (count($rows) === 0) {
            $response->getBody()->write('not found: isu');

            return $response->withStatus(StatusCodeInterface::STATUS_NOT_FOUND)
                ->withHeader('Content-Type', 'text/plain; charset=utf-8');
        }

        $res = Isu::fromDbRow($rows[0]);

        return $this->jsonResponse($response, $res);
    }

    /**
     * GET /api/isu/:jia_isu_uuid/icon
     * ISUのアイコンを取得
     */
    public function getIsuIcon(Request $request, Response $response, array $args): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $jiaIsuUuid = $args['jia_isu_uuid'];

        try {
            $stmt = $this->dbh->prepare('SELECT `image` FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?');
            $stmt->execute([$jiaUserId, $jiaIsuUuid]);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if (count($rows) === 0) {
            $response->getBody()->write('not found: isu');

            return $response->withStatus(StatusCodeInterface::STATUS_NOT_FOUND)
                    ->withHeader('Content-Type', 'text/plain; charset=utf-8');
        }

        $response->getBody()->write($rows[0]['image']);

        return $response;
    }

    /**
     * GET /api/isu/:jia_isu_uuid/graph
     * ISUのコンディショングラフ描画のための情報を取得
     */
    public function getIsuGraph(Request $request, Response $response, array $args): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $jiaIsuUuid = $args['jia_isu_uuid'];

        if (!isset($request->getQueryParams()['datetime'])) {
            $response->getBody()->write('missing: datetime');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $datetimeStr = $request->getQueryParams()['datetime'];
        $datetimeInt = filter_var($datetimeStr, FILTER_VALIDATE_INT);
        if (!is_int($datetimeInt)) {
            $response->getBody()->write('bad format: datetime');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $date = new DateTimeImmutable(date('Y-m-d H:00:00', $datetimeInt));

        $this->dbh->beginTransaction();

        try {
            $stmt = $this->dbh->prepare('SELECT COUNT(*) FROM `isu` WHERE `jia_user_id` = ? AND `jia_isu_uuid` = ?');
            $stmt->execute([$jiaUserId, $jiaIsuUuid]);
            $count = $stmt->fetch()[0];
        } catch (PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if ($count == 0) {
            $this->dbh->rollBack();
            $response->getBody()->write('not found: isu');

            return $response->withStatus(StatusCodeInterface::STATUS_NOT_FOUND)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        [$res, $err] = $this->generateIsuGraphResponse($jiaIsuUuid, $date);
        if (!empty($err)) {
            $this->dbh->rollBack();
            $this->logger->error($err);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        $this->dbh->commit();

        return $this->jsonResponse($response, $res);
    }

    /**
     * グラフのデータ点を一日分生成
     *
     * @return array{0: array<GraphResponse>, 1: string}
     */
    private function generateIsuGraphResponse(string $jiaIsuUuid, DateTimeImmutable $graphDate): array
    {
        /** @var array<GraphDataPointWithInfo> $dataPoints */
        $dataPoints = [];
        /** @var array<IsuCondition> $conditionsInThisHour */
        $conditionsInThisHour = [];
        /** @var array<int> $timestampsInThisHour */
        $timestampsInThisHour = [];
        /** @var DateTimeInterface $startTimeInThisHour */
        $startTimeInThisHour = (new DateTimeImmutable())->setTimestamp(0);

        try {
            $stmt = $this->dbh->prepare(
                'SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC',
                [PDO::ATTR_CURSOR => PDO::CURSOR_SCROLL],
            );
            $stmt->execute([$jiaIsuUuid]);

            while ($row = $stmt->fetch()) {
                $condition = IsuCondition::fromDbRow($row);

                $truncatedConditionTime = new DateTimeImmutable($condition->timestamp->format('Y-m-d H:00:00'));
                if ($truncatedConditionTime != $startTimeInThisHour) {
                    if (count($conditionsInThisHour) > 0) {
                        [$data, $err] = $this->calculateGraphDataPoint($conditionsInThisHour);
                        if (!empty($err)) {
                            return [[], $err];
                        }

                        $dataPoints[] = new GraphDataPointWithInfo(
                            jiaIsuUuid: $jiaIsuUuid,
                            startAt: $startTimeInThisHour,
                            data: $data,
                            conditionTimestamps: $timestampsInThisHour,
                        );
                    }

                    $startTimeInThisHour = $truncatedConditionTime;
                    $conditionsInThisHour = [];
                    $timestampsInThisHour = [];
                }

                $conditionsInThisHour[] = $condition;
                $timestampsInThisHour[] = $condition->timestamp->getTimestamp();
            }
        } catch (PDOException $e) {
            $err = 'db error: ' . $e->errorInfo[2];

            return [[], $err];
        }

        if (count($conditionsInThisHour) > 0) {
            [$data, $err] = $this->calculateGraphDataPoint($conditionsInThisHour);
            if (!empty($err)) {
                return [[], $err];
            }

            $dataPoints[] = new GraphDataPointWithInfo(
                jiaIsuUuid: $jiaIsuUuid,
                startAt: $startTimeInThisHour,
                data: $data,
                conditionTimestamps: $timestampsInThisHour,
            );
        }

        $endTime = $graphDate->modify('+24 hours');
        $startIndex = count($dataPoints);
        $endNextIndex = count($dataPoints);
        foreach ($dataPoints as $i => $graph) {
            if ($startIndex == count($dataPoints) && $graph->startAt >= $graphDate) {
                $startIndex = $i;
            }
            if ($endNextIndex == count($dataPoints) && $graph->startAt > $endTime) {
                $endNextIndex = $i;
            }
        }

        /** @var array<GraphDataPointWithInfo> $filteredDataPoints */
        $filteredDataPoints = [];
        if ($startIndex < $endNextIndex) {
            $filteredDataPoints = array_slice($dataPoints, $startIndex, $endNextIndex - $startIndex);
        }

        /** @var array<GraphResponse> $responseList */
        $responseList = [];
        $index = 0;
        $thisTime = $graphDate;

        while ($thisTime < $graphDate->modify('+24 hours')) {
            /** @var ?GraphDataPoint $data */
            $data = null;
            /** @var array<int> $timestamps */
            $timestamps = [];

            if ($index < count($filteredDataPoints)) {
                $dataWithInfo = $filteredDataPoints[$index];

                if ($dataWithInfo->startAt == $thisTime) {
                    $data = $dataWithInfo->data;
                    $timestamps = $dataWithInfo->conditionTimestamps;
                    $index++;
                }
            }

            $resp = new GraphResponse(
                startAt: $thisTime->getTimestamp(),
                endAt: $thisTime->modify('+1 hour')->getTimestamp(),
                data: $data,
                conditionTimestamps: $timestamps,
            );
            $responseList[] = $resp;

            $thisTime = $thisTime->modify('+1 hour');
        }

        return [$responseList, ''];
    }

    /**
     * 複数のISUのコンディションからグラフの一つのデータ点を計算
     *
     * @param array<IsuCondition> $isuConditions
     * @return array{0: GraphDataPoint, 1: string}
     */
    private function calculateGraphDataPoint(array $isuConditions): array
    {
        $conditionsCount = ['is_broken' => 0, 'is_dirty' => 0, 'is_overweight' => 0];
        $rawScore = 0;

        foreach ($isuConditions as $condition) {
            $badConditionsCount = 0;

            if (!$this->isValidConditionFormat($condition->condition)) {
                return [[], 'invalid condition format'];
            }

            foreach (explode(',', $condition->condition) as $condStr) {
                $keyValue = explode('=', $condStr);

                $conditionName = $keyValue[0];
                if ($keyValue[1] == 'true') {
                    $conditionsCount[$conditionName] += 1;
                    $badConditionsCount++;
                }
            }

            if ($badConditionsCount >= 3) {
                $rawScore += self::SCORE_CONDITION_LEVEL_CRITICAL;
            } elseif ($badConditionsCount >= 1) {
                $rawScore += self::SCORE_CONDITION_LEVEL_WARNING;
            } else {
                $rawScore += self::SCORE_CONDITION_LEVEL_INFO;
            }
        }

        $sittingCount = 0;
        foreach ($isuConditions as $condition) {
            if ($condition->isSitting) {
                $sittingCount++;
            }
        }

        $isuConditionsLength = count($isuConditions);

        $score = (int)($rawScore * 100 / 3 / $isuConditionsLength);

        $sittingPercentage = (int)($sittingCount * 100 / $isuConditionsLength);
        $isBrokenPercentage = (int)($conditionsCount['is_broken'] * 100 / $isuConditionsLength);
        $isOverweightPercentage = (int)($conditionsCount['is_overweight'] * 100 / $isuConditionsLength);
        $isDirtyPercentage = (int)($conditionsCount['is_dirty'] * 100 / $isuConditionsLength);

        $dataPoint = new GraphDataPoint(
            score: $score,
            percentage: new ConditionsPercentage(
                sitting: $sittingPercentage,
                isBroken: $isBrokenPercentage,
                isOverweight: $isOverweightPercentage,
                isDirty: $isDirtyPercentage,
            ),
        );

        return [$dataPoint, ''];
    }

    /**
     * GET /api/condition/:jia_isu_uuid
     * ISUのコンディションを取得
     */
    public function getIsuConditions(Request $request, Response $response, array $args): Response
    {
        [$jiaUserId, $errStatusCode, $err] = $this->getUserIdFromSession();

        if (!empty($err)) {
            $newResponse = $response->withStatus($errStatusCode);
            if ($errStatusCode === StatusCodeInterface::STATUS_UNAUTHORIZED) {
                $newResponse->getBody()->write('you are not signed in');

                return $newResponse->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }

            $this->logger->error($err);

            return $newResponse;
        }

        $jiaIsuUuid = $args['jia_isu_uuid'];
        if ($jiaIsuUuid === '') {
            $response->getBody()->write('missing: jia_isu_uuid');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        $params = $request->getQueryParams();
        if (!isset($params['end_time'])) {
            $response->getBody()->write('bad format: end_time');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $endTimeStr = $params['end_time'];
        $endTimeInt = filter_var($endTimeStr, FILTER_VALIDATE_INT);
        if (!is_int($endTimeInt)) {
            $response->getBody()->write('bad format: end_time');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $endTime = (new DateTimeImmutable())->setTimestamp($endTimeInt);

        if (!isset($params['condition_level'])) {
            $response->getBody()->write('missing: condition_level');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $conditionLevelCsv = $params['condition_level'];
        $conditionLevel = [];
        foreach (explode(',', $conditionLevelCsv) as $level) {
            $conditionLevel[$level] = [];
        }

        $startTimeStr = $params['start_time'] ?? '0';
        $startTimeInt = filter_var($startTimeStr, FILTER_VALIDATE_INT);
        if (!is_int($startTimeInt)) {
            $response->getBody()->write('bad format: start_time');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $startTime = (new DateTimeImmutable())->setTimestamp($startTimeInt);

        try {
            $stmt = $this->dbh->prepare('SELECT name FROM `isu` WHERE `jia_isu_uuid` = ? AND `jia_user_id` = ?');
            $stmt->execute([$jiaIsuUuid, $jiaUserId]);
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if (count($rows) === 0) {
            $response->getBody()->write('not found: isu');

            return $response->withStatus(StatusCodeInterface::STATUS_NOT_FOUND)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }
        $isuName = $rows[0]['name'];

        [$conditionsResponse, $err] = $this->getIsuConditionsFromDb(
            $jiaIsuUuid,
            $endTime,
            $conditionLevel,
            $startTime,
            self::CONDITION_LIMIT,
            $isuName,
        );
        if (!empty($err)) {
            $this->logger->error($err);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        return $this->jsonResponse($response, $conditionsResponse);
    }

    /**
     * ISUのコンディションをDBから取得
     *
     * @return array{0: array<GetIsuConditionResponse>, 1: string}
     */
    private function getIsuConditionsFromDb(
        string $jiaIsuUuid,
        DateTimeImmutable $endTime,
        array $conditionLevel,
        DateTimeImmutable $startTime,
        int $limit,
        string $isuName
    ): array {
        try {
            if ($startTime->getTimestamp() === 0) {
                $stmt = $this->dbh->prepare(
                    'SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?' .
                    '	AND `timestamp` < ?' .
                    '	ORDER BY `timestamp` DESC'
                );
                $stmt->execute([
                    $jiaIsuUuid,
                    $endTime->format('Y-m-d H:i:s'),
                ]);
            } else {
                $stmt = $this->dbh->prepare(
                    'SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ?' .
                    '	AND `timestamp` < ?' .
                    '	AND ? <= `timestamp`' .
                    '	ORDER BY `timestamp` DESC'
                );
                $stmt->execute([
                    $jiaIsuUuid,
                    $endTime->format('Y-m-d H:i:s'),
                    $startTime->format('Y-m-d H:i:s'),
                ]);
            }

            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $err = 'db error: ' . $e->errorInfo[2];

            return [[], $err];
        }

        /** @var array<IsuCondition> $conditions */
        $conditions = [];
        foreach ($rows as $row) {
            $conditions[] = IsuCondition::fromDbRow($row);
        }

        /** @var array<GetIsuConditionResponse> $conditionsResponse */
        $conditionsResponse = [];
        foreach ($conditions as $c) {
            try {
                $cLevel = $this->calculateConditionLevel($c->condition);
            } catch (UnexpectedValueException) {
                continue;
            }

            if (!isset($conditionLevel[$cLevel])) {
                continue;
            }

            $data = new GetIsuConditionResponse(
                jiaIsuUuid: $c->jiaIsuUuid,
                isuName: $isuName,
                timestamp: $c->timestamp->getTimestamp(),
                isSitting: $c->isSitting,
                condition: $c->condition,
                conditionLevel: $cLevel,
                message: $c->message,
            );
            $conditionsResponse[] = $data;
        }

        if (count($conditionsResponse) > $limit) {
            $conditionsResponse = array_slice($conditionsResponse, 0, $limit);
        }

        return [$conditionsResponse, ''];
    }

    /**
     * ISUのコンディションの文字列からコンディションレベルを計算
     *
     * @throws UnexpectedValueException
     */
    private function calculateConditionLevel(string $condition): string
    {
        $warnCount = mb_substr_count($condition, '=true');

        try {
            return match ($warnCount) {
                0 => self::CONDITION_LEVEL_INFO,
                1, 2 => self::CONDITION_LEVEL_WARNING,
                3 => self::CONDITION_LEVEL_CRITICAL,
            };
        } catch (UnhandledMatchError) {
            throw new UnexpectedValueException('unexpected warn count');
        }
    }

    /**
     * GET /api/trend
     * ISUの性格毎の最新のコンディション情報
     */
    public function getTrend(Request $request, Response $response): Response
    {
        try {
            $stmt = $this->dbh->query('SELECT `character` FROM `isu` GROUP BY `character`');
            $rows = $stmt->fetchAll();
        } catch (PDOException $e) {
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        /** @var array<Isu> $characterList */
        $characterList = [];
        foreach ($rows as $row) {
            $characterList[] = Isu::fromDbRow($row);
        }

        /** @var array<TrendResponse> $res */
        $res = [];
        foreach ($characterList as $character) {
            try {
                $stmt = $this->dbh->prepare('SELECT * FROM `isu` WHERE `character` = ?');
                $stmt->execute([$character->character]);
                $rows = $stmt->fetchAll();
            } catch (PDOException $e) {
                $this->logger->error('db error: ' . $e->errorInfo[2]);

                return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
            }

            /** @var array<Isu> $isuList */
            $isuList = [];
            foreach ($rows as $row) {
                $isuList[] = Isu::fromDbRow($row);
            }

            /** @var array<TrendCondition> $characterInfoIsuConditions */
            $characterInfoIsuConditions = [];
            /** @var array<TrendCondition> $characterWarningIsuConditions */
            $characterWarningIsuConditions = [];
            /** @var array<TrendCondition> $characterCriticalIsuConditions */
            $characterCriticalIsuConditions = [];

            foreach ($isuList as $isu) {
                try {
                    $stmt = $this->dbh->prepare('SELECT * FROM `isu_condition` WHERE `jia_isu_uuid` = ? ORDER BY timestamp DESC');
                    $stmt->execute([$isu->jiaIsuUuid]);
                    $rows = $stmt->fetchAll();
                } catch (PDOException $e) {
                    $this->logger->error('db error: ' . $e->errorInfo[2]);

                    return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
                }

                /** @var array<IsuCondition> $conditions */
                $conditions = [];
                foreach ($rows as $row) {
                    $conditions[] = IsuCondition::fromDbRow($row);
                }

                if (count($conditions) > 0) {
                    $isuLastCondition = $conditions[0];
                    try {
                        $conditionLevel = $this->calculateConditionLevel($isuLastCondition->condition);
                    } catch (UnexpectedValueException $e) {
                        $this->logger->error($e->getMessage());

                        return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
                    }

                    $trendCondition = new TrendCondition(
                        id: $isu->id,
                        timestamp: $isuLastCondition->timestamp->getTimestamp(),
                    );

                    switch ($conditionLevel) {
                        case 'info':
                            $characterInfoIsuConditions[] = $trendCondition;
                            break;
                        case 'warning':
                            $characterWarningIsuConditions[] = $trendCondition;
                            break;
                        case 'critical':
                            $characterCriticalIsuConditions[] = $trendCondition;
                            break;
                    }
                }
            }

            $cmp = function (TrendCondition $a, TrendCondition $b): int {
                return $b->timestamp <=> $a->timestamp;
            };
            usort($characterInfoIsuConditions, $cmp);
            usort($characterWarningIsuConditions, $cmp);
            usort($characterCriticalIsuConditions, $cmp);

            $res[] = new TrendResponse(
                character: $character->character,
                info: $characterInfoIsuConditions,
                warning: $characterWarningIsuConditions,
                critical: $characterCriticalIsuConditions,
            );
        }

        return $this->jsonResponse($response, $res);
    }

    /**
     * POST /api/condition/:jia_isu_uuid
     * ISUからのコンディションを受け取る
     */
    public function postIsuCondition(Request $request, Response $response, array $args): Response
    {
        // TODO: 一定割合リクエストを落としてしのぐようにしたが、本来は全量さばけるようにすべき
        $dropProbability = 0.9;

        if ((rand() / getrandmax()) <= $dropProbability) {
            $this->logger->warning('drop post isu condition request');

            return $response->withStatus(StatusCodeInterface::STATUS_ACCEPTED);
        }

        $jiaIsuUuid = $args['jia_isu_uuid'];
        if ($jiaIsuUuid === '') {
            $response->getBody()->write('missing: jia_isu_uuid');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        try {
            /** @var array<PostIsuConditionRequest> $req */
            $req = PostIsuConditionRequest::listFromJson((string)$request->getBody());
        } catch (UnexpectedValueException) {
            $response->getBody()->write('bad request body');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        if (count($req) === 0) {
            $response->getBody()->write('bad request body');

            return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        $this->dbh->beginTransaction();

        try {
            $stmt = $this->dbh->prepare('SELECT COUNT(*) FROM `isu` WHERE `jia_isu_uuid` = ?');
            $stmt->execute([$jiaIsuUuid]);
            $count = $stmt->fetch()[0];
        } catch (PDOException $e) {
            $this->dbh->rollBack();
            $this->logger->error('db error: ' . $e->errorInfo[2]);

            return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
        }

        if ($count == 0) {
            $this->dbh->rollBack();
            $response->getBody()->write('not found: isu');

            return $response->withStatus(StatusCodeInterface::STATUS_NOT_FOUND)
                ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
        }

        foreach ($req as $cond) {
            if (!$this->isValidConditionFormat($cond->condition)) {
                $this->dbh->rollBack();
                $response->getBody()->write('bad request body');

                return $response->withStatus(StatusCodeInterface::STATUS_BAD_REQUEST)
                    ->withHeader('Content-Type', 'text/plain; charset=UTF-8');
            }


            try {
                $stmt = $this->dbh->prepare(
                    'INSERT INTO `isu_condition`' .
                    '	(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`)' .
                    '	VALUES (?, ?, ?, ?, ?)'
                );
                $stmt->execute([
                    $jiaIsuUuid,
                    date('Y-m-d H:i:s', $cond->timestamp),
                    (int)$cond->isSitting,
                    $cond->condition,
                    $cond->message,
                ]);
            } catch (PDOException $e) {
                $this->dbh->rollBack();
                $this->logger->error('db error: ' . $e->errorInfo[2]);

                return $response->withStatus(StatusCodeInterface::STATUS_INTERNAL_SERVER_ERROR);
            }
        }

        $this->dbh->commit();

        return $response->withStatus(StatusCodeInterface::STATUS_ACCEPTED);
    }

    /**
     * ISUのコンディションの文字列がcsv形式になっているか検証
     */
    private function isValidConditionFormat(string $conditionStr): bool
    {
        $keys = ["is_dirty=", "is_overweight=", "is_broken="];

        $VALUE_TRUE = 'true';
        $VALUE_FALSE = 'false';

        $idxCondStr = 0;

        foreach ($keys as $idxKeys => $key) {
            if (!str_starts_with(mb_substr($conditionStr, $idxCondStr), $key)) {
                return false;
            }
            $idxCondStr += mb_strlen($key);

            if (str_starts_with(mb_substr($conditionStr, $idxCondStr), $VALUE_TRUE)) {
                $idxCondStr += mb_strlen($VALUE_TRUE);
            } elseif (str_starts_with(mb_substr($conditionStr, $idxCondStr), $VALUE_FALSE)) {
                $idxCondStr += mb_strlen($VALUE_FALSE);
            } else {
                return false;
            }

            if ($idxKeys < (count($keys) - 1)) {
                if ($conditionStr[$idxCondStr] !== ',') {
                    return false;
                }
                $idxCondStr++;
            }
        }

        return $idxCondStr == mb_strlen($conditionStr);
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

        $response->getBody()->write(file_get_contents($filePath));

        return $response->withHeader('Content-Type', $mimeType . '; charset=UTF-8');
    }

    /**
     * @throws UnexpectedValueException
     */
    private function jsonResponse(Response $response, JsonSerializable|array $data, int $statusCode = StatusCodeInterface::STATUS_OK): Response
    {
        $responseBody = json_encode($data, JSON_UNESCAPED_UNICODE);
        if ($responseBody === false) {
            throw new UnexpectedValueException('failed to json_encode');
        }

        $response->getBody()->write($responseBody);

        return $response->withStatus($statusCode)
            ->withHeader('Content-Type', 'application/json; charset=UTF-8');
    }
}
