import time
from django.core.cache import cache


def jwt_auth_middleware(get_response):
    def middleware(request):
        jwt_token = request.session.get('jwt_token')
        request.jwt_token = jwt_token
        request.is_logged_in = bool(jwt_token)
        request.is_admin = request.session.get('user_role') == 'admin'

        response = get_response(request)
        return response

    return middleware


def security_headers_middleware(get_response):
    def middleware(request):
        response = get_response(request)

         # HSTS - force HTTPS for 2 years
        response['Strict-Transport-Security'] = (
              "max-age=63072000; includeSubDomains; preload"
          )

         # X-Frame-Options - prevent clickjacking
        response['X-Frame-Options'] = 'DENY'

         # X-Content-Type-Options - prevent MIME sniffing
        response['X-Content-Type-Options'] = 'nosniff'

         # X-XSS-Protection - XSS filter
        response['X-XSS-Protection'] = '1; mode=block'

         # Referrer-Policy - control referrer info
        response['Referrer-Policy'] = 'strict-origin-when-cross-origin'

         # Permissions-Policy - restrict browser features
        response['Permissions-Policy'] = (
              "camera=(), microphone=(), geolocation=()"
          )

         # Remove server header when not in debug mode
        if request.META.get('HTTP_HOST', '').startswith('localhost'):
            response['Server'] = 'nginx'

        return response

    return middleware


def login_rate_limiter(get_response):
    def middleware(request):
        if (request.method == 'POST' 
            and request.path == '/login/'
            and not request.is_authenticated):
            x_forwarded_for = request.META.get('HTTP_X_FORWARDED_FOR', '')
            if x_forwarded_for:
                ip = x_forwarded_for.split(',')[0].strip()
            else:
                ip = request.META.get('REMOTE_ADDR', 'unknown')
            key = f"login_attempts:{ip}"
            attempts_data = cache.get(key)
            if attempts_data is None:
                attempts_data = {'count': 0, 'reset_time': time.time() + 300}
                cache.set(key, attempts_data, timeout=300)
            if attempts_data['count'] >= 5:
                if time.time() < attempts_data['reset_time']:
                    response = get_response(request)
                    response.status_code = 429
                    response['Retry-After'] = str(int(attempts_data['reset_time'] - time.time()))
                    response['Content-Type'] = 'text/plain'
                    response.reason_phrase = 'Too Many Requests'
                    return response
                else:
                    attempts_data = {'count': 0, 'reset_time': time.time() + 300}
            if request.method == 'POST':
                attempts_data['count'] = attempts_data.get('count', 0) + 1
                cache.set(key, attempts_data, timeout=300)
        return get_response(request)
    return middleware
