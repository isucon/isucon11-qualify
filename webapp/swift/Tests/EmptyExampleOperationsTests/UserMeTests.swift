//
// UserMeTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class UserMeTests: XCTestCase {

    func testUserMe() async throws {
        let input = UserMeRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleUserMe(input: input)
        XCTAssertEqual(response, UserMe200ResponseBody.__default)
    }
}
