//
// ListIsuGetTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class ListIsuGetTests: XCTestCase {

    func testListIsuGet() async throws {
        let input = ListIsuGetRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleListIsuGet(input: input)
        XCTAssertEqual(response, ListIsuGet200ResponseBody.__default)
    }
}
