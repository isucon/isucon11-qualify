<?php

declare(strict_types=1);

use Slim\App;
use Slim\Middleware\Session;
use App\Application\Middleware\AccessLog;

return function (App $app) {
    $app->add(AccessLog::class);
    $app->add(Session::class);
};
