// Package status provides constants for all standard HTTP status codes.
// HTTP status codes are returned by a server in response to a client's request to indicate
// whether the request was successfully processed, or if there were errors.
//
// These codes are divided into categories based on their response type:
//   - Informational responses (100–199): Indicates that the request was received and is being processed.
//   - Successful responses (200–299): Indicates that the request was successfully received, understood, and accepted.
//   - Redirection messages (300–399): Indicates that further action needs to be taken by the client to complete the request.
//   - Client error responses (400–499): Indicates that there was an error with the request sent by the client.
//   - Server error responses (500–599): Indicates that the server encountered an error while processing the request.
//
// Each status code is associated with a specific RFC section that defines its usage.
package status
