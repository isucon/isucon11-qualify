//
// Auth.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the Auth operation.

 - Parameters:
     - input: The validated AuthRequestBody object being passed to this operation.
     - context: The context provided for this operation.
 - Throws: badRequestBody, forbidden, internalServer.
 */
extension EmptyExampleOperationsContext {
    public func handleAuth(input: EmptyExampleModel.AuthRequestBody) async throws {
    }
}
