import 'dart:convert';
import 'dart:io';

enum Workspace { enterprise, department }

class Member {
  const Member({required this.organizationKind, required this.status});

  final String organizationKind;
  final String status;

  Workspace? get workspace => status != 'active'
      ? null
      : organizationKind == 'enterprise'
          ? Workspace.enterprise
          : organizationKind == 'department'
              ? Workspace.department
              : null;

  factory Member.fromJson(Map<String, dynamic> json) => Member(
        organizationKind: json['organizationKind'] as String? ?? '',
        status: json['status'] as String? ?? '',
      );
}

class Session {
  const Session({required this.member, this.accessToken, this.refreshToken});
  final Member member;
  final String? accessToken;
  final String? refreshToken;
}

abstract interface class TokenStore {
  Future<void> write(Session session);
  Future<void> clear();
}

class MemoryTokenStore implements TokenStore {
  Session? session;
  @override
  Future<void> write(Session value) async => session = value;
  @override
  Future<void> clear() async => session = null;
}

abstract interface class AuthClient {
  Future<Session> login(String loginId, String password);
  Future<void> logout(Session? session);
}

class HttpAuthClient implements AuthClient {
  HttpAuthClient(this.baseUri);
  final Uri baseUri;
  final HttpClient _client = HttpClient();

  @override
  Future<Session> login(String loginId, String password) async {
    final request = await _client.postUrl(baseUri.resolve('/v1/auth/login'));
    request.headers.contentType = ContentType.json;
    request.write(jsonEncode({
      'loginId': loginId,
      'password': password,
      'client': Platform.isAndroid ? 'flutter_android' : 'flutter_ios',
    }));
    final response = await request.close();
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw HttpException('login failed: ${response.statusCode}');
    }
    final body = jsonDecode(await response.transform(utf8.decoder).join()) as Map<String, dynamic>;
    return Session(
      member: Member.fromJson(body['member'] as Map<String, dynamic>),
      accessToken: body['accessToken'] as String?,
      refreshToken: body['refreshToken'] as String?,
    );
  }

  @override
  Future<void> logout(Session? session) async {
    final request = await _client.postUrl(baseUri.resolve('/v1/auth/logout'));
    if (session?.accessToken != null) {
      request.headers.set(HttpHeaders.authorizationHeader, 'Bearer ${session!.accessToken}');
    }
    final response = await request.close();
    if (response.statusCode != 204 &&
        (response.statusCode < 200 || response.statusCode >= 300)) {
      throw HttpException('logout failed: ${response.statusCode}');
    }
  }
}

class SessionController {
  SessionController(this.client, this.tokens);
  final AuthClient client;
  final TokenStore tokens;
  Session? session;

  Future<void> login(String loginId, String password) async {
    final next = await client.login(loginId, password);
    if (next.member.workspace == null) {
      await clear();
      throw StateError('no workspace');
    }
    session = next;
    await tokens.write(next);
  }

  Future<void> clear() async {
    session = null;
    await tokens.clear();
  }

  Future<void> logout() async {
    try {
      await client.logout(session);
    } finally {
      await clear();
    }
  }
}
