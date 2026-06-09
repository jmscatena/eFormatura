"""
finance_filters.py — Template filters customizados para o app finance.

Bug 5 fix: valores vindos da API Go são float (formato americano, ex: 15000.5).
Este filtro converte para o padrão monetário brasileiro (ex: 15.000,50).
"""

from django import template

register = template.Library()


@register.filter(name="brl")
def format_brl(value):
    """
    Formata um número float para o padrão monetário brasileiro.

    Uso no template: {{ income.amount|brl }}
    Saída: 1.500,00
    """
    try:
        return f"{float(value):,.2f}".replace(",", "X").replace(".", ",").replace("X", ".")
    except (ValueError, TypeError):
        return value
