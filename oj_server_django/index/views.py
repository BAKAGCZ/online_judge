from django.shortcuts import render


def index(request):
    return render(request, 'index/index.html')


def error(request, msg):
    context = {"msg": msg}
    return render(request, 'index/error.html', context=context)
