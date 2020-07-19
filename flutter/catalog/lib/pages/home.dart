import 'package:flutter/material.dart';

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: ConstrainedBox(
        constraints: const BoxConstraints.expand(),
        child: Center(
          child: RaisedButton(
            child: const Text('CONTINUE'),
            onPressed: () {
              Navigator.pushNamed(
                context,
                "/projects",
              );
            },
          ),
        ),
      ),
    );
  }
}
