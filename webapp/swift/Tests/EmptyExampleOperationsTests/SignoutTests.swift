//
// SignoutTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class SignoutTests: XCTestCase {

    func testSignout() async throws {
        let input = SignoutRequestBody.__default
        let operationsContext = createOperationsContext()
    
        try await operationsContext.handleSignout(input: input)
    }
}
