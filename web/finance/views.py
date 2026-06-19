import requests
import jwt
import json
import time
from django.shortcuts import render, redirect
from django.contrib import messages
from django.urls import reverse
from django.core.exceptions import PermissionDenied
from django.http import HttpResponse, HttpResponseNotAllowed, HttpResponseForbidden, HttpResponseBadRequest

import os

API_URL = os.environ.get("API_URL", "http://127.0.0.1:8080/api")
AUTH_URL = os.environ.get("AUTH_URL", "http://127.0.0.1:8080")
API_TIMEOUT = int(os.environ.get("API_TIMEOUT", "5"))


def get_auth_headers(request):
    return {"Authorization": f"Bearer {request.jwt_token}"} if request.jwt_token else {}


def extract_paginated_data(response_json):
    """Extrai a lista de dados de uma resposta paginada do backend.
    
    O backend retorna: {"data": [...], "page": 1, "limit": 50, "total": N, "total_pages": M}
    Esta função retorna apenas a lista em "data", ou a resposta original se já for uma lista.
    """
    if isinstance(response_json, dict) and "data" in response_json:
        return response_json["data"] or []
    if isinstance(response_json, list):
        return response_json
    return []


def api_request_with_retry(method, url, **kwargs):
    """Faz requisição HTTP com retry logic e exponential backoff."""
    timeout = kwargs.pop("timeout", API_TIMEOUT)
    retries = int(os.environ.get("API_RETRIES", "3"))
    
    for attempt in range(retries):
        try:
            resp = method(url, **kwargs, timeout=timeout)
            return resp
        except requests.exceptions.RequestException:
            if attempt == retries - 1:
                raise
            time.sleep(2 ** attempt)


def admin_required(view_func):
    def wrapper(request, *args, **kwargs):
        if not request.is_admin:
            messages.error(
                request, "Acesso negado. Você não tem permissões de administrador."
            )
            return redirect("dashboard")
        return view_func(request, *args, **kwargs)

    return wrapper


def login_view(request):
    if request.method == "POST":
        email = request.POST.get("email")
        password = request.POST.get("password")
        try:
            resp = api_request_with_retry(requests.post,
                f"{AUTH_URL}/login",
                json={"email": email, "password": password},
                timeout=API_TIMEOUT,
            )
            if resp.status_code == 200:
                token = resp.json().get("token")
                # Decodificar sem verificar a assinatura — o Django confia na API Go que emitiu o token
                decoded = jwt.decode(token, options={"verify_signature": False})
                request.session["jwt_token"] = token
                request.session["user_role"] = decoded.get("role", "comum")
                return redirect("dashboard")
            else:
                try:
                    err_msg = resp.json().get("error", "Credenciais inválidas.")
                except Exception:
                    err_msg = f"Erro no login (Status {resp.status_code})"
                messages.error(request, err_msg)
        except Exception as e:
            import logging
            logging.exception("Erro no login_view:")
            messages.error(request, f"Erro interno ao realizar login: {e}")

    return render(request, "finance/login.html")


def logout_view(request):
    request.session.flush()
    return redirect("login")


def dashboard(request):
    if not request.is_logged_in:
        return redirect("login")

    try:
        headers = get_auth_headers(request)
        incomes_resp = api_request_with_retry(requests.get, f"{API_URL}/incomes", headers=headers, timeout=API_TIMEOUT)
        expenses_resp = api_request_with_retry(requests.get, f"{API_URL}/expenses", headers=headers, timeout=API_TIMEOUT)

        incomes = extract_paginated_data(incomes_resp.json()) if incomes_resp.status_code == 200 else []
        expenses = extract_paginated_data(expenses_resp.json()) if expenses_resp.status_code == 200 else []

        total_arrecadado = sum(inc.get("amount", 0) for inc in incomes)

        # Bug 1 fix: custos operacionais usam total_amount (valor comprometido),
        # não apenas parcelas pagas.
        # Bug 2 fix: parcelas pagas de contratos-alvo também são descontadas dos
        # fundos livres, pois representam dinheiro efetivamente gasto.
        total_custo_operacional = 0   # Custos operacionais (comprometidos)
        total_contratos_meta = 0      # Valor total dos contratos-alvo (meta)
        total_contratos_pagos = 0     # Parcelas pagas de contratos-alvo

        for exp in expenses:
            cat = exp.get("category", "Contrato")
            if cat == "Custo":
                # Custo operacional: deduz o valor total comprometido, não só o pago
                total_custo_operacional += exp.get("total_amount", 0)
            else:
                # Contrato-alvo: acumula a meta e rastreia o que já foi pago
                total_contratos_meta += exp.get("total_amount", 0)
                for inst in exp.get("installments", []):
                    if inst.get("status") == "Pago":
                        total_contratos_pagos += inst.get("amount", 0)

        # Fundos livres = arrecadado - custos operacionais - parcelas de contratos pagas
        total_descontado = total_custo_operacional + total_contratos_pagos
        fundos_livres = total_arrecadado - total_descontado

        # Bug 3 fix: progresso mede quanto dos contratos-alvo já foi pago,
        # não fundos_livres / meta (que era semanticamente incorreto)
        if total_contratos_meta > 0:
            progresso_meta = (total_contratos_pagos / total_contratos_meta) * 100
        else:
            progresso_meta = 0

        progresso_meta = max(0, min(100, progresso_meta))

        def fmt_brl(value):
            """Formata um float para o padrão monetário brasileiro."""
            return f"{value:,.2f}".replace(",", "X").replace(".", ",").replace("X", ".")

        fmt_arrecadado = fmt_brl(total_arrecadado)
        fmt_custo = fmt_brl(total_custo_operacional)
        fmt_contratos_pagos = fmt_brl(total_contratos_pagos)
        fmt_total_descontado = fmt_brl(total_descontado)
        fmt_fundos = fmt_brl(fundos_livres)
        fmt_meta = fmt_brl(total_contratos_meta)
        fmt_porcentagem = int(progresso_meta)

    except Exception as e:
        fmt_arrecadado = "0,00"
        fmt_custo = "0,00"
        fmt_contratos_pagos = "0,00"
        fmt_total_descontado = "0,00"
        fmt_fundos = "0,00"
        fmt_meta = "0,00"
        fmt_porcentagem = 0

    context = {
        "total_arrecadado": fmt_arrecadado,
        "total_custo_operacional": fmt_custo,
        # Detalhes do desconto para transparência no template
        "total_contratos_pagos": fmt_contratos_pagos,
        "total_descontado": fmt_total_descontado,
        "fundos_livres": fmt_fundos,
        "total_contratos_meta": fmt_meta,
        "progresso_meta": fmt_porcentagem,
        "is_admin": request.is_admin,
    }
    return render(request, "finance/dashboard.html", context)


def incomes_list(request):
    if not request.is_logged_in:
        return redirect("login")
    headers = get_auth_headers(request)
    incomes = []
    try:
        resp = api_request_with_retry(requests.get, f"{API_URL}/incomes", headers=headers, timeout=API_TIMEOUT)
        if resp.status_code == 200:
            incomes = extract_paginated_data(resp.json())
    except requests.exceptions.RequestException:
        pass

    # Bug 4 fix: calcular o total no backend em vez de concatenar no template
    total_incomes = sum(inc.get("amount", 0) for inc in incomes)
    fmt_total_incomes = (
        f"{total_incomes:,.2f}".replace(",", "X").replace(".", ",").replace("X", ".")
    )

    return render(
        request,
        "finance/incomes_list.html",
        {
            "incomes": incomes,
            "total_incomes": fmt_total_incomes,
            "is_admin": request.is_admin,
        },
    )


@admin_required
def income_create(request):
    if request.method == "POST":
        amount_str = request.POST.get("amount", "0")
        try:
            amount = float(amount_str)
        except (ValueError, TypeError):
            amount = 0
            messages.error(request, "Valor inválido para a receita.")
        data = {
            "title": request.POST.get("title"),
            "amount": amount,
            "category": request.POST.get("category"),
        }
        api_request_with_retry(requests.post,
            f"{API_URL}/incomes", json=data, headers=get_auth_headers(request), timeout=API_TIMEOUT
        )
        return redirect("incomes_list")
    return render(request, "finance/form_income.html")


@admin_required
def income_delete(request, id):
    if request.method == "POST":
        try:
            resp = api_request_with_retry(requests.delete, f"{API_URL}/incomes/{id}", headers=get_auth_headers(request), timeout=API_TIMEOUT)
            if resp.status_code not in (200, 204):
                messages.error(request, f"Erro ao excluir: {resp.json().get('error', 'Erro desconhecido')}")
            else:
                messages.success(request, "Receita exclu\u00edda com sucesso!")
        except Exception as e:
            messages.error(request, f"Erro ao excluir: {str(e)}")
    return redirect("incomes_list")


def notifications_list(request):
    if not request.is_logged_in:
        return redirect("login")

    headers = get_auth_headers(request)
    notifications = []

    try:
        resp = api_request_with_retry(requests.get, f"{API_URL}/notifications", headers=headers, timeout=API_TIMEOUT)
        if resp.status_code == 200:
            notifications = extract_paginated_data(resp.json())
    except requests.exceptions.RequestException:
        pass

    return render(
        request,
        "finance/notifications.html",
        {"notifications": notifications, "is_admin": request.is_admin},
    )


def expenses_list(request):
    if not request.is_logged_in:
        return redirect("login")
    headers = get_auth_headers(request)
    expenses = []
    try:
        resp = api_request_with_retry(requests.get, f"{API_URL}/expenses", headers=headers, timeout=API_TIMEOUT)
        if resp.status_code == 200:
            expenses = extract_paginated_data(resp.json())
    except requests.exceptions.RequestException:
        pass
    return render(
        request,
        "finance/expenses_list.html",
        {"expenses": expenses, "is_admin": request.is_admin},
    )


@admin_required
def expense_create(request):
    if request.method == "POST":
        installments = request.POST.get("installment_count", "1")
        try:
            installment_count = int(installments)
        except ValueError:
            installment_count = 1

        total_amount_str = request.POST.get("total_amount", "0")
        try:
            total_amount = float(total_amount_str)
        except (ValueError, TypeError):
            total_amount = 0
            messages.error(request, "Valor inválido para o contrato.")

        data = {
            "contract_name": request.POST.get("contract_name"),
            "description": request.POST.get("description"),
            "total_amount": total_amount,
            "installment_count": installment_count,
            "category": request.POST.get("category", "Contrato"),
            "start_date": request.POST.get("start_date", ""),
            "first_payment_date": request.POST.get("first_payment_date", ""),
        }
        api_request_with_retry(requests.post,
            f"{API_URL}/expenses", json=data, headers=get_auth_headers(request), timeout=API_TIMEOUT
        )
        return redirect("expenses_list")
    return render(request, "finance/form_expense.html")


@admin_required
def expense_delete(request, id):
    if request.method == "POST":
        try:
            resp = api_request_with_retry(requests.delete, f"{API_URL}/expenses/{id}", headers=get_auth_headers(request), timeout=API_TIMEOUT)
            if resp.status_code not in (200, 204):
                messages.error(request, f"Erro ao excluir: {resp.json().get('error', 'Erro desconhecido')}")
            else:
                messages.success(request, "Contrato exclu\u00eddo com sucesso!")
        except Exception as e:
            messages.error(request, f"Erro ao excluir: {str(e)}")
    return redirect("expenses_list")


@admin_required
def installment_pay(request, id):
    if request.method == "POST":
        try:
            resp = api_request_with_retry(requests.put,
                f"{API_URL}/installments/{id}/pay",
                headers=get_auth_headers(request),
                timeout=API_TIMEOUT
             )
            if resp.status_code == 200:
                messages.success(request, "Parcela paga com sucesso!")
            else:
                messages.error(request, f"Erro ao pagar: {resp.json().get('error', 'Erro desconhecido')}")
        except Exception as e:
            messages.error(request, f"Erro ao pagar: {str(e)}")
    return redirect("expenses_list")


@admin_required
def user_create(request):
    if request.method == "POST":
        data = {
             "name": request.POST.get("name"),
             "email": request.POST.get("email"),
             "password": request.POST.get("password"),
             "role": "comum",
         }
        try:
            resp = api_request_with_retry(requests.post, f"{AUTH_URL}/register", json=data, timeout=API_TIMEOUT)
            if resp.status_code == 201:
                messages.success(request, "Usuário criado com sucesso!")
                return redirect("dashboard")
            else:
                messages.error(
                    request,
                    f"Erro ao criar usuário: {resp.json().get('error', 'Erro desconhecido')}",
                 )
        except Exception as e:
            messages.error(request, f"Erro de conexão: {str(e)}")

    return render(request, "finance/form_user.html")


@admin_required
def users_list(request):
    if request.method == "GET":
        return render_users_list(request)

    # Handle POST actions (update, reset_password, disable)
    user_id_str = request.POST.get("user_id", "0")
    try:
        user_id = int(user_id_str)
    except (ValueError, TypeError):
        user_id = 0
        messages.error(request, "ID de usuário inválido.")
        return redirect("users_list")
    action = request.POST.get("action")

    if action == "update":
        data = {
              "name": request.POST.get("name"),
              "email": request.POST.get("email"),
          }
        resp = api_request_with_retry(requests.put,
            f"{API_URL}/users/{user_id}",
            json=data,
            headers=get_auth_headers(request),
            timeout=API_TIMEOUT,
          )
        if resp.status_code == 200:
            messages.success(request, "Usuário atualizado com sucesso!")
        else:
            messages.error(request, f"Erro ao atualizar: {resp.json().get('error', 'Erro desconhecido')}")
        return redirect("users_list")

    elif action == "reset_password":
        password = request.POST.get("password")
        confirm = request.POST.get("confirm_password")
        if password != confirm:
            messages.error(request, "As senhas não coincidem.")
            return redirect("users_list")
        elif len(password) < 8:
            messages.error(request, "A senha deve ter pelo menos 8 caracteres.")
            return redirect("users_list")
        else:
            data = {"password": password}
            resp = api_request_with_retry(requests.put,
                f"{API_URL}/users/{user_id}/password",
                json=data,
                headers=get_auth_headers(request),
                timeout=API_TIMEOUT,
              )
            if resp.status_code == 200:
                messages.success(request, "Senha redefinida com sucesso!")
            else:
                messages.error(request, f"Erro ao redefinir: {resp.json().get('error', 'Erro desconhecido')}")
            return redirect("users_list")

    elif action == "disable":
        disabled = request.POST.get("disabled") == "true"
        data = {"disabled": disabled}
        resp = api_request_with_retry(requests.put,
            f"{API_URL}/users/{user_id}/disable",
            json=data,
            headers=get_auth_headers(request),
            timeout=API_TIMEOUT,
          )
        if resp.status_code == 200:
            status = "Desativado" if disabled else "Reativado"
            messages.success(request, f"Usuário {status} com sucesso!")
        else:
            messages.error(request, f"Erro ao atualizar status: {resp.json().get('error', 'Erro desconhecido')}")
        return redirect("users_list")

    return redirect("users_list")


def render_users_list(request):
    headers = get_auth_headers(request)
    users = []
    try:
        resp = api_request_with_retry(requests.get, f"{API_URL}/users", headers=headers, timeout=API_TIMEOUT)
        if resp.status_code == 200:
            users = extract_paginated_data(resp.json())
    except requests.exceptions.RequestException:
        pass

    total = len(users)
    active = sum(1 for u in users if not u.get("disabled", False))
    disabled = total - active

    return render(
        request,
        "finance/users_list.html",
        {
            "users": users,
            "total_users": total,
            "active_users": active,
            "disabled_users": disabled,
            "is_admin": request.is_admin,
        },
    )


def webhook_notifications(request):
    """Recebe notificações do backend Go via webhook."""
    if request.method != "POST":
        return HttpResponseNotAllowed(["POST"])
    
    # Verificar autenticação do webhook
    webhook_secret = request.META.get('HTTP_X_WEBHOOK_SECRET', '')
    if webhook_secret != os.environ.get('WEBHOOK_SECRET', ''):
        return HttpResponseForbidden("Webhook secret invalid")
    
    try:
        data = json.loads(request.body)
        # Salvar notificação no banco ou processar imediatamente
        title = data.get('title', 'Notificação')
        messages.info(request, f"Notificação: {title}")
        return HttpResponse(status=200)
    except Exception as e:
        return HttpResponseBadRequest(str(e))
