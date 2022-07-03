//
// GetCustomerDetails.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the GetCustomerDetails operation.

 - Parameters:
     - input: The validated GetCustomerDetailsRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The CustomerAttributes object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: unknownResource.
 */
extension EmptyExampleOperationsContext {
    public func handleGetCustomerDetails(input: EmptyExampleModel.GetCustomerDetailsRequest) async throws
    -> EmptyExampleModel.CustomerAttributes {
        return CustomerAttributes.__default
    }
}
