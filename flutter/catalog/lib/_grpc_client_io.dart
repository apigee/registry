import 'package:grpc/grpc.dart' as grpc;
import 'dart:io' show Platform;

String token; // auth token

grpc.ClientChannel createClientChannel() {
  Map<String, String> env = Platform.environment;
  token = env['CLI_FLAME_TOKEN'];
  final insecure = env['CLI_FLAME_INSECURE'];
  final address = env['CLI_FLAME_ADDRESS'];
  final parts = address.split(":");
  final host = parts[0];
  final port = int.parse(parts[1]);
  final channelOptions = (insecure == "1")
      ? const grpc.ChannelOptions(
          credentials: const grpc.ChannelCredentials.insecure())
      : const grpc.ChannelOptions(
          credentials: const grpc.ChannelCredentials.secure());
  return grpc.ClientChannel(host, port: port, options: channelOptions);
}

grpc.CallOptions callOptions() {
  if (token == null) {
    return grpc.CallOptions();
  }
  Map<String, String> metadata = {"authorization": "Bearer " + token};
  grpc.CallOptions callOptions = grpc.CallOptions(metadata: metadata);
  return callOptions;
}
