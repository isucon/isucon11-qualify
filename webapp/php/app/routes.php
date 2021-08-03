<?php
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
};

function getIndex(Request $request, Response $response): Response {
    $response->getBody()->write(file_get_contents(FRONTEND_CONTENTS_PATH . '/index.html'));

    return $response;
}
