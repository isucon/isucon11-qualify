//
// IsuGraph.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the IsuGraph operation.

 - Parameters:
     - input: The validated IsuGraphRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The IsuGraph200ResponseBody object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: badRequestBody, internalServer, unauthorized, unknownResource.
 */
extension EmptyExampleOperationsContext {
    public func handleIsuGraph(input: EmptyExampleModel.IsuGraphRequest) async throws
    -> EmptyExampleModel.IsuGraph200ResponseBody {
        return IsuGraph200ResponseBody.__default
    }
}
