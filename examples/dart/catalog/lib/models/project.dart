import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';

extension Display on Project {
  String nameForDisplay() {
    if (this.displayName != "") {
      return this.displayName;
    } else {
      return this.name;
    }
  }

  String routeNameForProjectDetail() {
    final name = "/" + this.name.split("/").sublist(1).join("/");
    print("pushing " + name);
    return name;
  }
}
