from django.urls import path
from . import views

urlpatterns = [
    path("login/", views.login_view, name="login"),
    path("logout/", views.logout_view, name="logout"),
    path("register/", views.user_create, name="register"),
    path("", views.dashboard, name="dashboard"),
    path("incomes/", views.incomes_list, name="incomes_list"),
    path("incomes/create/", views.income_create, name="income_create"),
    path("incomes/<int:id>/delete/", views.income_delete, name="income_delete"),
    path("expenses/", views.expenses_list, name="expenses_list"),
    path("expenses/create/", views.expense_create, name="expense_create"),
    path("expenses/<int:id>/delete/", views.expense_delete, name="expense_delete"),
    path("installments/<int:id>/pay/", views.installment_pay, name="installment_pay"),
    path("notifications/", views.notifications_list, name="notifications_list"),
    path("api/webhook/notifications/", views.webhook_notifications, name="webhook_notifications"),
    path("users/", views.users_list, name="users_list"),
]
