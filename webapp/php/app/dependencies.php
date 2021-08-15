<?php

declare(strict_types=1);

use App\Application\Middleware\AccessLog as AccessLogMiddleware;
use App\Application\Settings\SettingsInterface;
use DI\ContainerBuilder;
use GuzzleHttp\Client as HttpClient;
use GuzzleHttp\ClientInterface as HttpClientInterface;
use Monolog\Handler\StreamHandler;
use Monolog\Logger;
use Monolog\Processor\UidProcessor;
use Psr\Container\ContainerInterface;
use Psr\Log\LoggerInterface;
use Slim\Middleware\Session;
use SlimSession\Helper as SessionHelper;

return function (ContainerBuilder $containerBuilder) {
    $containerBuilder->addDefinitions([
        AccessLogMiddleware::class => function (ContainerInterface $c): AccessLogMiddleware {
            $logger = new Logger('access-log');

            $handler = new StreamHandler('php://stdout');
            $logger->pushHandler($handler);

            return new AccessLogMiddleware($logger);
        },
        HttpClientInterface::class => function (ContainerInterface $c): HttpClientInterface {
            return new HttpClient();
        },
        LoggerInterface::class => function (ContainerInterface $c): LoggerInterface {
            $settings = $c->get(SettingsInterface::class);

            $loggerSettings = $settings->get('logger');
            $logger = new Logger($loggerSettings['name']);

            $processor = new UidProcessor();
            $logger->pushProcessor($processor);

            $handler = new StreamHandler($loggerSettings['path'], $loggerSettings['level']);
            $logger->pushHandler($handler);

            return $logger;
        },
        PDO::class => function (ContainerInterface $c): PDO {
            $databaseSettings = $c->get(SettingsInterface::class)->get('database');

            $dsn = vsprintf('mysql:host=%s;dbname=%s;port=%d;charset=utf8mb4', [
                $databaseSettings['host'],
                $databaseSettings['database'],
                $databaseSettings['port']
            ]);

            $pdo = new PDO($dsn, $databaseSettings['user'], $databaseSettings['password'], [
                // TODO: 持続的な接続で問題ないか要検証
                PDO::ATTR_PERSISTENT => true,
            ]);

            $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

            return $pdo;
        },
        Session::class => function (ContainerInterface $c): Session {
            $sessionSettings = $c->get(SettingsInterface::class)->get('session');

            return new Session($sessionSettings);
        },
        SessionHelper::class => function (ContainerInterface $c): SessionHelper {
            return new SessionHelper();
        }
    ]);
};
