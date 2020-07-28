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

import 'dart:convert';
import 'package:archive/archive.dart';
import 'package:flutter/material.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';
import 'package:catalog/generated/third_party/gnostic/metrics/complexity.pb.dart';
import '../service/service.dart';
import '../application.dart';

class SpecDetailWidget extends StatefulWidget {
  final Spec spec;
  final String name;

  SpecDetailWidget(this.spec, this.name);
  @override
  _SpecDetailWidgetState createState() => _SpecDetailWidgetState(this.spec);
}

String projectNameForSpecName(String specName) {
  final parts = specName.split("/");
  return parts.sublist(0, 2).join("/");
}

class _SpecDetailWidgetState extends State<SpecDetailWidget> {
  Spec spec;
  List<Property> properties;
  String body;

  _SpecDetailWidgetState(this.spec);

  @override
  Widget build(BuildContext context) {
    final specName = "projects" + widget.name;
    if (spec == null) {
      // we need to fetch the spec from the API
      final specFuture = SpecService.getSpec(specName);
      specFuture.then((spec) {
        setState(() {
          this.spec = spec;
          if ((spec.contents != null) && (spec.contents.length > 0)) {
            final data = GZipDecoder().decodeBytes(spec.contents);
            this.body = Utf8Codec().decoder.convert(data);
            print(this.body);
          }
        });
        print(spec);
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

    if (properties == null) {
      // fetch the properties
      final propertiesFuture = PropertiesService.listProperties(
          projectNameForSpecName(specName),
          subject: specName);
      propertiesFuture.then((properties) {
        setState(() {
          this.properties = properties.properties;
        });
        print(properties);
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
        child: Column(children: [
          Row(children: [specCard(context, spec)]),
          if (propertiesContain(properties, "summary"))
            Row(children: [
              summaryCard(context, spec, properties),
            ]),
          Row(children: [Text(body != null ? body : "")]),
        ]),
      ),
    );
  }
}

Expanded specCard(BuildContext context, Spec spec) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title:
                Text(spec.name, style: Theme.of(context).textTheme.headline5),
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

bool propertiesContain(List<Property> properties, String propertyName) {
  if (properties == null) {
    return false;
  }
  for (var p in properties) {
    if (p.relation == propertyName) return true;
  }
  return false;
}

Property propertyWithName(List<Property> properties, String propertyName) {
  if (properties == null) {
    return null;
  }
  for (var p in properties) {
    if (p.relation == propertyName) return p;
  }
  return null;
}

TableRow row(BuildContext context, String label, String value) {
  return TableRow(
    children: [
      Padding(
        padding: EdgeInsets.all(5),
        child: Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
      Padding(
        padding: EdgeInsets.all(5),
        child: Text(
          value,
          textAlign: TextAlign.center,
        ),
      ),
    ],
  );
}

Expanded summaryCard(
    BuildContext context, Spec spec, List<Property> properties) {
  final summary = propertyWithName(properties, "complexity");
  Complexity complexitySummary =
      new Complexity.fromBuffer(summary.messageValue.value);
  print("$summary");
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              Expanded(child: SizedBox()),
              Expanded(
                child: Table(
                  border: TableBorder.all(),
                  children: [
                    //TableRow(children: [Text("Objects")]),
                    row(context, "Paths", "${complexitySummary.pathCount}"),
                    row(context, "Schemas", "${complexitySummary.schemaCount}"),
                    row(context, "Schema Properties",
                        "${complexitySummary.schemaPropertyCount}"),
                  ],
                ),
              ),
              Expanded(child: SizedBox()),
              Expanded(
                child: Table(
                  border: TableBorder.all(),
                  children: [
                    // TableRow(children: [Text("Operations")]),
                    row(context, "Operations",
                        "${complexitySummary.getCount + complexitySummary.postCount + complexitySummary.putCount + complexitySummary.deleteCount}"),
                    row(context, "GETs", "${complexitySummary.getCount}"),
                    row(context, "POSTs", "${complexitySummary.postCount}"),
                    row(context, "PUTs", "${complexitySummary.putCount}"),
                    row(context, "DELETEs", "${complexitySummary.deleteCount}"),
                  ],
                ),
              ),
              Expanded(child: SizedBox()),
            ],
          ),
        ],
      ),
    ),
  );
}
