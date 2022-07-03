//
// EmptyExampleTestConfiguration.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel
import Logging

struct TestVariables {
    static let logger = Logger(label: "EmptyExampleTestConfiguration")
}

func createOperationsContext() -> EmptyExampleOperationsContext {
    return EmptyExampleOperationsContext(logger: TestVariables.logger)
}
