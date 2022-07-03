//
// UserMe.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the UserMe operation.

 - Parameters:
     - input: The validated UserMeRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The UserMe200ResponseBody object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: internalServer, unauthorized.
 */
extension EmptyExampleOperationsContext {
    public func handleUserMe(input: EmptyExampleModel.UserMeRequest) async throws
    -> EmptyExampleModel.UserMe200ResponseBody {
        return UserMe200ResponseBody.__default
    }
}
