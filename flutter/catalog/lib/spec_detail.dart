import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';
import 'package:catalog/generated/flame_properties.pb.dart';

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
        });
        print(spec);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            "API Hub",
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
            "API Hub",
          ),
        ),
        body: Text("loading..."),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Hub",
        ),
      ),
      body: SingleChildScrollView(
        child: Column(children: [
          Row(children: [specCard(context, spec)]),
          if (propertiesContain(properties, "summary"))
            Row(children: [summaryCard(context, spec, properties)]),
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
            subtitle: Text("$spec"),
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

Expanded summaryCard(
    BuildContext context, Spec spec, List<Property> properties) {
  final summary = propertyWithName(properties, "summary");
  ComplexitySummary complexitySummary =
      new ComplexitySummary.fromBuffer(summary.messageValue.value);
  print("$complexitySummary");
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text("$complexitySummary"),
        ],
      ),
    ),
  );
}
