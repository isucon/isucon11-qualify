<?php

// phpcs:disable PSR1.Files.SideEffects
declare(strict_types=1);

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\App;

return function (App $app) {
    $app->options('/{routes:.*}', function (Request $request, Response $response) {
        // CORS Pre-Flight OPTIONS Request Handler
        return $response;
    });

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
