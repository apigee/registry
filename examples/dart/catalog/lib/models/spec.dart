import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';

extension Display on Spec {
  String nameForDisplay() {
    if (this.name != "") {
      return this.name;
    } else {
      return this.filename;
    }
  }

  String routeNameForSpecDetail() {
    final name = "/" + this.name.split("/").sublist(1).join("/");
    print("pushing " + name);
    return name;
  }
}
