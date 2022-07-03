//
// IsuGraphTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class IsuGraphTests: XCTestCase {

    func testIsuGraph() async throws {
        let input = IsuGraphRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleIsuGraph(input: input)
        XCTAssertEqual(response, IsuGraph200ResponseBody.__default)
    }
}
