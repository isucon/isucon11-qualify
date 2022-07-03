//
// ListIsuGet.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the ListIsuGet operation.

 - Parameters:
     - input: The validated ListIsuGetRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The ListIsuGet200ResponseBody object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: internalServer, unauthorized.
 */
extension EmptyExampleOperationsContext {
    public func handleListIsuGet(input: EmptyExampleModel.ListIsuGetRequest) async throws
    -> EmptyExampleModel.ListIsuGet200ResponseBody {
        return ListIsuGet200ResponseBody.__default
    }
}
