import 'package:flutter/material.dart';
import 'package:catalog/generated/registry_models.pb.dart';
import '../service/service.dart';
import '../application.dart';
import '../models/projects.dart';

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
    if (project.description != null) {
      return project.description;
    }
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
          title: Text(
            applicationName,
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
        title: Text(applicationName),
      ),
      body: SingleChildScrollView(
        child: Center(
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
                        title: Text(project.nameForDisplay(),
                            style: Theme.of(context).textTheme.headline2),
                        subtitle: Text(subtitlePropertyText()),
                      ),
                      ButtonBar(
                        children: <Widget>[
                          FlatButton(
                            child: const Text('APIS',
                                semanticsLabel: "APIs BUTTON"),
                            onPressed: () {
                              Navigator.pushNamed(
                                context,
                                routeNameForProjectDetailProducts(project),
                                arguments: project,
                              );
                            },
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
