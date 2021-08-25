
func (c *RegistryClient) GrpcClient() rpcpb.RegistryClient {
	return c.internalClient.(*registryGRPCClient).registryClient
}
