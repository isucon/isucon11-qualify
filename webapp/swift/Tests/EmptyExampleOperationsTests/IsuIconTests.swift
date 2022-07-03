//
// IsuIconTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class IsuIconTests: XCTestCase {

    func testIsuIcon() async throws {
        let input = IsuIconRequest.__default
        let operationsContext = createOperationsContext()
    
        try await operationsContext.handleIsuIcon(input: input)
    }
}
