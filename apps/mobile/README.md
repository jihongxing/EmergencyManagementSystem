# Mobile

One Flutter application for Android and iOS. Enterprise and administrative
workspaces remain unavailable until server-side authentication and authorization
are implemented. The current screen is an unauthenticated engineering shell.

From this directory:

```powershell
flutter pub get --enforce-lockfile
flutter analyze --no-pub
flutter test --no-pub
```

Native Android/iOS scaffolding and the dependency lockfile are present.
Analysis and widget tests have passed on Windows. APK and iOS builds, signing,
device testing and remote CI have not been validated. Generated application
identifiers are development placeholders, not approved distribution identifiers.

See [engineering governance](../../docs/engineering.md) for checks and
process-scoped proxy guidance. Do not commit local proxy settings or credentials.
