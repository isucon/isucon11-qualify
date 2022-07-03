//
// CreateCustomerPut.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the CreateCustomerPut operation.

 - Parameters:
     - input: The validated CreateCustomerRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The CreateCustomerPut200Response object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: unknownResource.
 */
extension EmptyExampleOperationsContext {
    public func handleCreateCustomerPut(input: EmptyExampleModel.CreateCustomerRequest) async throws
    -> EmptyExampleModel.CreateCustomerPut200Response {
        return CreateCustomerPut200Response.__default
    }
}
