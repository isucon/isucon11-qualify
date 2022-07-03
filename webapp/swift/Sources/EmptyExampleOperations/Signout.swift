//
// Signout.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the Signout operation.

 - Parameters:
     - input: The validated SignoutRequestBody object being passed to this operation.
     - context: The context provided for this operation.
 - Throws: internalServer.
 */
extension EmptyExampleOperationsContext {
    public func handleSignout(input: EmptyExampleModel.SignoutRequestBody) async throws {
    }
}
