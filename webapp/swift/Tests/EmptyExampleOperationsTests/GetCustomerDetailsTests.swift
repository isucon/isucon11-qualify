//
// GetCustomerDetailsTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class GetCustomerDetailsTests: XCTestCase {

    func testGetCustomerDetails() async throws {
        let input = GetCustomerDetailsRequest.__default
        let operationsContext = createOperationsContext()
    
        let response = try await operationsContext.handleGetCustomerDetails(input: input)
        XCTAssertEqual(response, CustomerAttributes.__default)
    }
}
