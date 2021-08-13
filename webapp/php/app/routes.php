<?php

// phpcs:disable PSR1.Files.SideEffects
declare(strict_types=1);

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\App;

const FRONTEND_CONTENTS_PATH = __DIR__ . '/../../public';

return function (App $app) {
    $app->options('/{routes:.*}', function (Request $request, Response $response) {
        // CORS Pre-Flight OPTIONS Request Handler
        return $response;
    });

    $app->get('/', 'getIndex');
    $app->get('/condition', 'getIndex');
    $app->get('/isu/{jia_isu_uuid}', 'getIndex');
    $app->get('/register', 'getIndex');
    $app->get('/login', 'getIndex');

    $app->get('/assets/{filename}', 'getAssets');
};

function getIndex(Request $request, Response $response): Response
{
    $response->getBody()->write(file_get_contents(FRONTEND_CONTENTS_PATH . '/index.html'));

    return $response;
}

function getAssets(Request $request, Response $response, array $args): Response
{
    $filePath = FRONTEND_CONTENTS_PATH . '/assets/' . $args['filename'];

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
