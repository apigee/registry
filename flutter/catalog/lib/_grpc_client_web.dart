import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;
import 'package:grpc/grpc.dart' as grpc;

// web app needs an openly-available test server
String url = "https://registry-backend-3rqz64w4vq-uw.a.run.app";

String token;

GrpcWebClientChannel createClientChannel() =>
    GrpcWebClientChannel.xhr(Uri.parse(url));

grpc.CallOptions callOptions() {
  if (token == null) {
    return grpc.CallOptions();
  }
  Map<String, String> metadata = {"Authorization": "Bearer " + token};
  grpc.CallOptions callOptions = grpc.CallOptions(metadata: metadata);
  return callOptions;
}