<?php

declare(strict_types=1);

use Slim\App;
use Slim\Middleware\Session;

return function (App $app) {
    $app->add(Session::class);
};
