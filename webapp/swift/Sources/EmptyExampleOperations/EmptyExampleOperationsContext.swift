//
// EmptyExampleOperationsContext.swift
// EmptyExampleOperations
//

import Foundation
import Logging

/**
 The context to be passed to each of the EmptyExample operations.
 */
public struct EmptyExampleOperationsContext {
    let logger: Logger
    // TODO: Add properties to be accessed by the operation handlers

    public init(logger: Logger) {
        self.logger = logger
    }
}
