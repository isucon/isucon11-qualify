//
// IsuConditionGetTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class IsuConditionGetTests: XCTestCase {

    func testIsuConditionGet() async throws {
        let input = IsuConditionGetRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleIsuConditionGet(input: input)
        XCTAssertEqual(response, IsuConditionGet200ResponseBody.__default)
    }
}
