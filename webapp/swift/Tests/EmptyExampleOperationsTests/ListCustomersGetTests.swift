//
// ListCustomersGetTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class ListCustomersGetTests: XCTestCase {

    func testListCustomersGet() async throws {
        let input = ListCustomersGetRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleListCustomersGet(input: input)
        XCTAssertEqual(response, ListCustomersResponse.__default)
    }
}
