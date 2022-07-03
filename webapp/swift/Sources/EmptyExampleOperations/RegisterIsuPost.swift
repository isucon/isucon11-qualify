//
// RegisterIsuPost.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the RegisterIsuPost operation.

 - Parameters:
     - input: The validated RegisterIsuPostRequestBody object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The IsuAttributes object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: badRequestBody, conflict, internalServer, unauthorized.
 */
extension EmptyExampleOperationsContext {
    public func handleRegisterIsuPost(input: EmptyExampleModel.RegisterIsuPostRequestBody) async throws
    -> EmptyExampleModel.IsuAttributes {
        return IsuAttributes.__default
    }
}
