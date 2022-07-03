//
// EmptyExamplePerInvocationContextInitializer.swift
// EmptyExampleService
//

import EmptyExampleOperations
import EmptyExampleOperationsHTTP1
import SmokeOperationsHTTP1Server
import SmokeAWSCore
import NIO
            
/**
 Initializer for the EmptyExampleService.
 */
@main
struct EmptyExamplePerInvocationContextInitializer: EmptyExamplePerInvocationContextInitializerProtocol {
    // TODO: Add properties to be accessed by the operation handlers

    /**
     On application startup.
     */
    init(eventLoopGroup: EventLoopGroup) async throws {
        CloudwatchStandardErrorLogger.enableLogging()

        // TODO: Add additional application initialization
    }

    /**
     On invocation.
    */
    public func getInvocationContext(invocationReporting: SmokeServerInvocationReporting<SmokeInvocationTraceContext>)
    -> EmptyExampleOperationsContext {
        return EmptyExampleOperationsContext(logger: invocationReporting.logger)
    }

    /**
     On application shutdown.
    */
    func onShutdown() async throws {
        
    }
}
