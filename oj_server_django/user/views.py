from django.shortcuts import render
from django.http import HttpResponseRedirect as redirect
from django.http import HttpResponse
from django.contrib import auth
from django.contrib.auth import get_user_model
from django.contrib.auth.models import User


def userinfo(request):
    user = request.user
    Users = get_user_model()
    point_total = Users.objects.get(username=user.username).point
    rank = Users.objects.filter(point__gt=point_total).count()+1
    context = {'user': user, 'rank': rank}
    return render(request, 'user/info.html', context=context)


def ranking(request):
    User = get_user_model()
    context = {'user': User.objects.order_by("-point")[0:100]}
    return render(request, 'user/ranking.html', context=context)


def login(request):
    return render(request, 'user/login.html')


def loginhandler(request):
    if request.method == 'POST':
        username = request.POST.get('username')
        password = request.POST.get('password')
    # 内置验证
    user = auth.authenticate(username=username, password=password)
    if user:  # 登录信息、session封装到request.user
        auth.login(request, user)
        return redirect('/')
    context = {'msg': "登陆失败！请重试"}
    return render(request, 'index/error.html', context=context)


def logout(request):
    auth.logout(request)
    return redirect('/')


def register(request):
    return render(request, "user/register.html")


def registerhandler(request):
    if request.method == 'POST':
        username = request.POST.get('username')
        password = request.POST.get('password')
    User = get_user_model()
    if User.objects.filter(username=username).exists():
        context = {'msg': "用户名已存在，请重试！"}
        return render(request, 'index/error.html', context=context)
    User.objects.create_user(
        username=username, password=password, point=0, attempt_number=0, solved_number=0)
    return HttpResponse("注册成功！<a href='/'>回到主页</a>")
