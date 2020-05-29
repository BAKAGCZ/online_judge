from django.shortcuts import render
from django.http import HttpResponseRedirect as redirect
from django.http import HttpResponse
from django.contrib import auth
from django.contrib.auth import get_user_model
from django.contrib.auth.models import User
from django.core.paginator import Paginator


def userinfo(request):
    user = request.user
    users = get_user_model()
    point_total = users.objects.get(username=user.username).point
    rank = users.objects.filter(point__gt=point_total).count()+1
    context = {'user': user, 'rank': rank}
    return render(request, 'user/info.html', context=context)


def ranking(request, pindex):
    User = get_user_model()
    user_list = User.objects.order_by("-point")
    paginator = Paginator(user_list, 5)
    if pindex == "":  # django默认返回空值，设置默认值1
        pindex = 1
    else:  # 如果有返回值，把返回值转为整数型
        int(pindex)
    page = paginator.page(pindex)  # 传递当前页的实例对象到前端
    context = {'page': page, 'type': 'list'}
    request.encoding = 'utf-8'
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
    if username == None or username == "":
        context = {'msg': "注册失败！请重试"}
        return render(request, 'index/error.html', context=context)
    if len(username) > 50:
        context = {'msg': "用户名过长，请重试！"}
        return render(request, 'index/error.html', context=context)
    if len(password) > 50:
        context = {'msg': "密码过长，请重试！"}
        return render(request, 'index/error.html', context=context)
    if len(password) < 6:
        context = {'msg': "密码过短，请重试！"}
        return render(request, 'index/error.html', context=context)
    User = get_user_model()
    if User.objects.filter(username=username).exists():
        context = {'msg': "用户名已存在，请重试！"}
        return render(request, 'index/error.html', context=context)
    User.objects.create_user(
        username=username, password=password, point=0, attempt_number=0, solved_number=0)
    context = {'msg': "注册成功！"}
    return render(request, 'index/msg.html', context=context)
