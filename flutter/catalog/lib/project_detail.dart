import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';
import 'drawer.dart';

class ProjectDetailWidget extends StatefulWidget {
  final Project project;
  final String name;

  ProjectDetailWidget(this.project, this.name);
  @override
  _ProjectDetailWidgetState createState() =>
      _ProjectDetailWidgetState(this.project);
}

String routeNameForProjectDetailProducts(Project project) {
  final name = "/" + project.name.split("/").sublist(1).join("/") + "/products";
  print("pushing " + name);
  return name;
}

class _ProjectDetailWidgetState extends State<ProjectDetailWidget> {
  Project project;
  List<Property> properties;

  _ProjectDetailWidgetState(this.project);

  String subtitlePropertyText() {
    if (properties == null) {
      return "";
    }
    for (var property in properties) {
      if (property.relation == "subtitle") {
        return property.stringValue;
      }
    }
    return "";
  }

  @override
  Widget build(BuildContext context) {
    final projectName = "projects" + widget.name;
    if (project == null) {
      // we need to fetch the project from the API
      final projectFuture = ProjectService.getProject(projectName);
      projectFuture.then((project) {
        setState(() {
          this.project = project;
        });
        print(project);
      });
      return Scaffold(
        appBar: AppBar(
          title: Row(
            children: [
              Text("left"),
              Text(
                "API Hub",
              ),
            ],
          ),
        ),
        body: Text("loading..."),
      );
    }

    if (properties == null) {
      // fetch the properties
      final propertiesFuture =
          PropertiesService.listProperties(projectName, subject: projectName);
      propertiesFuture.then((properties) {
        setState(() {
          this.properties = properties.properties;
        });
        print(properties);
      });
    }

    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            IconButton(
              icon: const Icon(Icons.arrow_back_ios),
              onPressed: () {
                Navigator.pop(context);
              },
              tooltip: MaterialLocalizations.of(context).backButtonTooltip,
            ),
            Text(
              "API Hub",
            ),
          ],
        ),
        leading: Builder(
          builder: (BuildContext context) {
            return IconButton(
              icon: const Icon(Icons.menu),
              onPressed: () {
                Scaffold.of(context).openDrawer();
              },
              tooltip: MaterialLocalizations.of(context).openAppDrawerTooltip,
            );
          },
        ),
      ),
      drawer: drawer(context),
      body: SingleChildScrollView(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            minHeight: 10,
            maxHeight: 800,
          ),
          child: Container(
            decoration: BoxDecoration(
                //color:Colors.yellow,
                ),
            margin: EdgeInsets.fromLTRB(40, 20, 40, 0),
            padding: EdgeInsets.fromLTRB(0, 0, 0, 0),
            child: Column(
              children: [
                Card(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: <Widget>[
                      ListTile(
                        title: Text(project.name.split("/").last,
                            style: Theme.of(context).textTheme.headline1),
                        subtitle: Text(subtitlePropertyText()),
                      ),
                      ButtonBar(
                        children: <Widget>[
                          FlatButton(
                            child: const Text('API PRODUCTS'),
                            onPressed: () {
                              Navigator.pushNamed(
                                context,
                                routeNameForProjectDetailProducts(project),
                                arguments: project,
                              );
                            },
                          ),
                          FlatButton(
                            child: const Text('APPLICATIONS'),
                            onPressed: () {/* ... */},
                          ),
                          FlatButton(
                            child: const Text('TEAMS'),
                            onPressed: () {/* ... */},
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
