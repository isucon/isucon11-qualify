//
// InitializeTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class InitializeTests: XCTestCase {

    func testInitialize() async throws {
        let input = InitializeRequestBody.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleInitialize(input: input)
        XCTAssertEqual(response, InitializeAttributes.__default)
    }
}
