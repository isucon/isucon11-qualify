//
// Trend.swift
// EmptyExampleOperations
//

import Foundation
import EmptyExampleModel

/**
 Handler for the Trend operation.

 - Parameters:
     - input: The validated TrendRequest object being passed to this operation.
     - context: The context provided for this operation.
 - Returns: The Trend200ResponseBody object to be passed back from the caller of this operation.
     Will be validated before being returned to caller.
 - Throws: internalServer.
 */
extension EmptyExampleOperationsContext {
    public func handleTrend(input: EmptyExampleModel.TrendRequest) async throws
    -> EmptyExampleModel.Trend200ResponseBody {
        return Trend200ResponseBody.__default
    }
}
