/*
Package drip provides networking and state synchronization for the Bappa Framework.

Drip enables server-client communication with custom state serialization
and transmission. The server runs on a fixed timestep and handles core game logic,
while clients receive state updates and can send input to the server.

Currently in early development. Only supports networking for a single scene. Uses a basic server authoritative
TCP architecture.
*/
package drip
