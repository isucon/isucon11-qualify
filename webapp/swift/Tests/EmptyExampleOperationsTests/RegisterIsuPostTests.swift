//
// RegisterIsuPostTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class RegisterIsuPostTests: XCTestCase {

    func testRegisterIsuPost() async throws {
        let input = RegisterIsuPostRequestBody.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleRegisterIsuPost(input: input)
        XCTAssertEqual(response, IsuAttributes.__default)
    }
}
