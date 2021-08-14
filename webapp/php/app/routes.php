<?php

// phpcs:disable PSR1.Files.SideEffects,PSR1.Classes.ClassDeclaration
declare(strict_types=1);

use Fig\Http\Message\StatusCodeInterface;
use Firebase\JWT\JWT;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Log\LoggerInterface;
use Slim\App;
use SlimSession\Helper as SessionHelper;

return function (App $app) {
    $app->options('/{routes:.*}', function (Request $request, Response $response) {
        // CORS Pre-Flight OPTIONS Request Handler
        return $response;
    });

    $app->post('/api/auth', Handler::class . ':postAuthentication');
    $app->post('/api/signout', Handler::class . ':postSignout');

    $app->get('/', Handler::class . ':getIndex');
    $app->get('/condition', Handler::class . ':getIndex');
    $app->get('/isu/{jia_isu_uuid}', Handler::class . ':getIndex');
    $app->get('/register', Handler::class . ':getIndex');
    $app->get('/login', Handler::class . ':getIndex');

    $app->get('/assets/{filename}', Handler::class . ':getAssets');
};

final class Handler
{
    private const FRONTEND_CONTENTS_PATH = __DIR__ . '/../../public';
    private const JIA_JWT_SIGNING_KEY_PATH = __DIR__ . '/../../ec256-public.pem';

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

        // TODO: 公開鍵の検証は必要？

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
}
