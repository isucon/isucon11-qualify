<?php

declare(strict_types=1);

namespace App\Application\Middleware;

use DateTimeInterface;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Http\Server\RequestHandlerInterface as RequestHandler;
use Psr\Log\LoggerInterface;

class AccessLog
{
    public function __construct(
        private LoggerInterface $logger
    ) {
    }

    public function __invoke(Request $request, RequestHandler $handler): Response
    {
        $start = hrtime(true);

        $response = $handler->handle($request);

        $serverParams = $request->getServerParams();
        $host = $request->getUri()->getHost();
        if (!is_null($port = $request->getUri()->getPort())) {
            $host .= ':' . $port;
        }

        $this->logger->info('', [
            'time' => date(DatetimeInterface::RFC3339, $serverParams['REQUEST_TIME'] ?? null),
            'remote_ip' => $request->getServerParams()['REMOTE_ADDR'] ?? '-',
            'host' => $host,
            'method' => $request->getMethod(),
            'uri' => $request->getRequestTarget(),
            'status' => $response->getStatusCode(),
            'latency' => hrtime(true) - $start,
            'bytes_in' => $request->getBody()->getSize() ?? '-',
            'bytes_out' => $response->getBody()->getSize() ?? '-',
        ]);

        return $response;
    }
}
