import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;

final url = "http://localhost:9999";

GrpcWebClientChannel createClientChannel() =>
    GrpcWebClientChannel.xhr(Uri.parse(url));
