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

class ApiDetailWidget extends StatefulWidget {
  final Api api;
  final String name;

  ApiDetailWidget(this.api, this.name);
  @override
  _ApiDetailWidgetState createState() => _ApiDetailWidgetState(this.api);
}

String routeNameForApiDetailVersions(Api api) {
  final name = "/" + api.name.split("/").sublist(1).join("/") + "/versions";
  print("pushing " + name);
  return name;
}

class _ApiDetailWidgetState extends State<ApiDetailWidget> {
  Api api;
  List<Property> properties;

  _ApiDetailWidgetState(this.api);

  @override
  Widget build(BuildContext context) {
    final apiName = "projects" + widget.name;
    if (api == null) {
      // we need to fetch the api from the API
      final apiFuture = ApiService.getApi(apiName);
      apiFuture.then((api) {
        setState(() {
          this.api = api;
        });
        print(api);
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
              Row(children: [apiCard(context, api)]),
              Row(children: [
                apiInfoCard(context, api),
                apiInfoCard(context, api)
              ]),
              Row(children: [
                apiInfoCard(context, api),
                apiInfoCard(context, api),
                apiInfoCard(context, api)
              ]),
            ],
          ),
        ),
      ),
    );
  }
}

Expanded apiCard(BuildContext context, Api api) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title: Text(api.displayName,
                style: Theme.of(context).textTheme.headline5),
            subtitle: Text(api.description),
          ),
          ButtonBar(
            children: <Widget>[
              FlatButton(
                child: const Text('VERSIONS'),
                onPressed: () {
                  Navigator.pushNamed(
                    context,
                    routeNameForApiDetailVersions(api),
                    arguments: api,
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

Expanded apiInfoCard(BuildContext context, Api api) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title:
                Text("More Info", style: Theme.of(context).textTheme.headline6),
            subtitle: Text("$api"),
          ),
          ButtonBar(
            children: <Widget>[
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
