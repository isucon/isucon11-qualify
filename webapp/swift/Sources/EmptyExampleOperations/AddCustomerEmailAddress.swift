//
// AddCustomerEmailAddress.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the AddCustomerEmailAddress operation.

 - Parameters:
     - input: The validated AddCustomerEmailAddressRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The CustomerEmailAddressIdentity object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: concurrency, customerEmailAddressAlreadyExists, customerEmailAddressLimitExceeded, unknownResource.
 */
extension EmptyExampleOperationsContext {
    public func handleAddCustomerEmailAddress(input: EmptyExampleModel.AddCustomerEmailAddressRequest) async throws
    -> EmptyExampleModel.CustomerEmailAddressIdentity {
        return CustomerEmailAddressIdentity.__default
    }
}
