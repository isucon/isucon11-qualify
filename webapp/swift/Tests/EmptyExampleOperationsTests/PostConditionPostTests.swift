//
// PostConditionPostTests.swift
// EmptyExampleOperationsTests
//

import XCTest
@testable import EmptyExampleOperations
import EmptyExampleModel

class PostConditionPostTests: XCTestCase {

    func testPostConditionPost() async throws {
        let input = PostConditionPostRequest.__default
        let operationsContext = createOperationsContext()
    
        try await operationsContext.handlePostConditionPost(input: input)
    }
}
