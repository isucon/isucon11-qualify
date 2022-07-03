//
// AuthTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class AuthTests: XCTestCase {

    func testAuth() async throws {
        let input = AuthRequestBody.__default
        let operationsContext = createOperationsContext()
    
        try await operationsContext.handleAuth(input: input)
    }
}
