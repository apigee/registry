const {ListProductsRequest, ListProductsResponse} = require('./flame_service_pb.js');
const {FlameClient} = require('./flame_service_grpc_web_pb.js');

var client = new FlameClient('http://localhost:9999');

var request = new ListProductsRequest();
request.setParent('projects/google');
request.setPageSize(3);

console.log(request);

client.listProducts(request, {}, (err, response) => {
  console.log(err);
  console.log(response);
});
