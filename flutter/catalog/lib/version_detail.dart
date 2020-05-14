import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

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
        child: ConstrainedBox(
          constraints: BoxConstraints(
            minHeight: 200,
            maxHeight: 800,
          ),
          child: Column(
            children: [
              versionCard(context, version),
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
                style: Theme.of(context).textTheme.headline4),
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
