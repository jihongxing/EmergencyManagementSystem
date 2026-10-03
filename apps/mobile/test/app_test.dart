import 'package:emergency_mobile/auth/session.dart';
import 'package:emergency_mobile/main.dart';
import 'package:flutter_test/flutter_test.dart';

class FakeAuthClient implements AuthClient {
  FakeAuthClient(this.member);
  final Member member;
  @override
  Future<Session> login(String loginId, String password) async => Session(member: member);
  @override
  Future<void> logout(Session? session) async {}
}

void main() {
  testWidgets('unauthenticated app exposes login only', (tester) async {
    final auth = SessionController(
      FakeAuthClient(const Member(organizationKind: 'enterprise', status: 'active')),
      MemoryTokenStore(),
    );
    await tester.pumpWidget(EmergencyApp(auth: auth));
    expect(find.text('登录'), findsOneWidget);
    expect(find.text('企业工作区'), findsNothing);
    expect(find.text('行政工作区'), findsNothing);
  });

  testWidgets('server identity selects workspace and logout clears it', (tester) async {
    final auth = SessionController(
      FakeAuthClient(const Member(organizationKind: 'department', status: 'active')),
      MemoryTokenStore(),
    );
    await tester.pumpWidget(EmergencyApp(auth: auth));
    await tester.tap(find.text('登录'));
    await tester.pump();
    expect(find.text('行政工作区'), findsOneWidget);
    expect(find.text('组织购买（占位）'), findsNothing);
    await tester.tap(find.text('退出登录'));
    await tester.pump();
    expect(find.text('行政工作区'), findsNothing);
    expect(find.text('登录'), findsOneWidget);
  });
}
