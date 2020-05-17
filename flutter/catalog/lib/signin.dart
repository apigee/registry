import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'authorizations.dart';

GoogleSignInAccount currentUser;
bool currentUserIsAuthorized = false;

GoogleSignIn googleSignIn = GoogleSignIn(
  scopes: <String>[
    'email',
  ],
);

Future<GoogleSignInAccount> attemptToSignIn() async {
  googleSignIn.onCurrentUserChanged.listen((GoogleSignInAccount account) {
    currentUser = account;
    currentUserIsAuthorized = authorizedUsers.contains(currentUser.email);
    print("signed in: $currentUser (authorized = $currentUserIsAuthorized)");
  });
  return googleSignIn.signInSilently();
}

class SignInScreen extends StatefulWidget {
  @override
  State createState() => SignInScreenState();
}

class SignInScreenState extends State<SignInScreen> {
  @override
  void initState() {
    super.initState();
    googleSignIn.onCurrentUserChanged.listen((GoogleSignInAccount account) {
      setState(() {});
    });
  }

  Future<void> _handleSignIn() async {
    try {
      await googleSignIn.signIn();
    } catch (error) {
      print(error);
    }
  }

  Future<void> _handleSignOut() => googleSignIn.disconnect();

  Widget _buildBody(BuildContext context) {
    if (currentUser != null) {
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              GoogleUserCircleAvatar(
                identity: currentUser,
              ),
              Container(
                margin: EdgeInsets.fromLTRB(20, 0, 20, 0),
                child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(currentUser.displayName ?? '',
                          textAlign: TextAlign.left,
                          style: Theme.of(context).textTheme.headline6),
                      Text(currentUser.email ?? '',
                          textAlign: TextAlign.left,
                          style: Theme.of(context).textTheme.bodyText1),
                    ]),
              ),
              RaisedButton(
                child: const Text('SIGN OUT'),
                onPressed: _handleSignOut,
              ),
            ],
          ),
          Container(height: 30),
          Text("API Hub is an early-stage prototype."),
          Container(height: 10),
          Text("For information, contact govlife-team@google.com."),
          Container(height: 10),
          if (currentUserIsAuthorized)
            RaisedButton(
              child: const Text('CONTINUE'),
              onPressed: () {
                Navigator.pushNamed(
                  context,
                  "/projects",
                );
              },
            ),
        ],
      );
    } else {
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Text("API Hub", style: Theme.of(context).textTheme.headline2),
          RaisedButton(
            child: const Text('SIGN IN WITH GOOGLE'),
            onPressed: _handleSignIn,
          ),
          Container(height: 20),
          Text("For evaluation only."),
        ],
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: ConstrainedBox(
        constraints: const BoxConstraints.expand(),
        child: _buildBody(context),
      ),
    );
  }
}
