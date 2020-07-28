// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import 'package:flutter/material.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';
import '../service/service.dart';
import '../application.dart';

extension StringSplitting on String {
  String last(int n) {
    final parts = this.split("/");
    final sublist = parts.sublist(parts.length - 2 * n, parts.length - 0);
    return sublist.join("/");
  }
}

class VersionDetailWidget extends StatefulWidget {
  final Version version;
  final String name;

  VersionDetailWidget(this.version, this.name);
  @override
  _VersionDetailWidgetState createState() =>
      _VersionDetailWidgetState(this.version);
}

String routeNameForVersionDetailSpecs(Version version) {
  final name = "/" + version.name.split("/").sublist(1).join("/") + "/specs";
  print("pushing " + name);
  return name;
}

class _VersionDetailWidgetState extends State<VersionDetailWidget> {
  Version version;
  List<Property> properties;

  _VersionDetailWidgetState(this.version);

  @override
  Widget build(BuildContext context) {
    final versionName = "projects" + widget.name;
    if (version == null) {
      // we need to fetch the version from the API
      final versionFuture = VersionService.getVersion(versionName);
      versionFuture.then((version) {
        setState(() {
          this.version = version;
        });
        print(version);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            applicationName,
          ),
        ),
        body: Text("loading..."),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: Text(
          applicationName,
        ),
      ),
      body: SingleChildScrollView(
        child: Center(
          child: Column(
            children: [
              Row(children: [versionCard(context, version)]),
            ],
          ),
        ),
      ),
    );
  }
}

Expanded versionCard(BuildContext context, Version version) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title: Text(version.name,
                style: Theme.of(context).textTheme.headline5),
            subtitle: Text("$version"),
          ),
          ButtonBar(
            children: <Widget>[
              FlatButton(
                child: const Text('SPECS'),
                onPressed: () {
                  Navigator.pushNamed(
                    context,
                    routeNameForVersionDetailSpecs(version),
                    arguments: version,
                  );
                },
              ),
              FlatButton(
                child: const Text('MORE'),
                onPressed: () {/* ... */},
              ),
            ],
          ),
        ],
      ),
    ),
  );
}
