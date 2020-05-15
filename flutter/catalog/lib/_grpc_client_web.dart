import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;
import 'package:grpc/grpc.dart' as grpc;


String url = "http://localhost:9999";
String token; // auth token

GrpcWebClientChannel createClientChannel() =>
    GrpcWebClientChannel.xhr(Uri.parse(url));

grpc.CallOptions callOptions() {
  if (token == null) {
    return grpc.CallOptions();
  }
  Map<String, String> metadata = {"authorization": "Bearer " + token};
  grpc.CallOptions callOptions = grpc.CallOptions(metadata: metadata);
  return callOptions;
}