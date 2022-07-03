//
// GetIsu.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the GetIsu operation.

 - Parameters:
     - input: The validated GetIsuRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The IsuAttributes object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: internalServer, unauthorized, unknownResource.
 */
extension EmptyExampleOperationsContext {
    public func handleGetIsu(input: EmptyExampleModel.GetIsuRequest) async throws
    -> EmptyExampleModel.IsuAttributes {
        return IsuAttributes.__default
    }
}
