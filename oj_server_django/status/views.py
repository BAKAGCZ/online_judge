from django.shortcuts import render
from django.shortcuts import redirect
# from django.urls import reverse
# from django.http import HttpResponseRedirect
from websocket import create_connection
from problem.models import Problem
from .models import Solution
from django.contrib.auth import get_user_model
import datetime
import json
from django.core.paginator import Paginator


def submithandler(request):
    request.encoding = 'utf-8'
    username = request.user.username

    if request.method == 'POST':
        # problemResObj = Problem.objects
        # if problem_id != 0:
        #     solution_res_obj = solution_res_obj.filter(problem_id=problem_id)
        problem_id_int = int(request.POST['problem_id'])
        problem_id = str(problem_id_int)
        lang = request.POST['lang']
        code = request.POST['code']
        submitted = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    # 题号不存在
    if not Problem.objects.filter(id=problem_id).exists():
        context = {'msg': "题目不存在！"}
        return render(request, 'index/error.html', context=context)
    problem = Problem.objects.get(id=problem_id)
    time_limit = problem.time_limit
    memory_limit = problem.memory_limit

    if username == None or username == "":
        return render(request, 'user/login.html')
    elif problem_id == None or problem_id == "" or lang == None or lang == "" or code == None or code == "":
        context = {'msg': "数据无效！"}
        return render(request, 'index/error.html', context=context)
    else:
        send_params = {
            "ProblemID": problem_id,
            "Username": username,
            "Code": code,
            "Submitted": submitted,
            "Lang": lang,
            "TimeLimit": time_limit,
            "MemoryLimit": memory_limit,
        }

        try:
            address = "ws://127.0.0.1:8886/websocket"
            ws = create_connection(address)
            ws.send(json.dumps(send_params))  # 将字典形式的数据转化为字符串
            # print("Sent")
            # print("Receiving...")
            recv_params = json.loads(ws.recv())
            solution = Solution()
            solution.username = recv_params['Username']
            solution.problem_id = recv_params['ProblemID']
            solution.result = recv_params['Result']
            solution.memory = recv_params['Memory']
            solution.time = recv_params['Time']
            solution.lang = recv_params['Lang']
            solution.length = recv_params['Length']
            solution.submitted = recv_params['Submitted']
            solution.code = recv_params['Code']

            User = get_user_model()
            user = User.objects.get(username=username)
            if solution.result == "Accepted":
                # 没解决过
                if not Solution.objects.filter(username=username, problem_id=problem_id, result=solution.result).exists():
                    user.point += problem.point
                    user.solved_number += 1
            user.attempt_number += 1

            solution.save()
            user.save()

            # print("Received: "+recv_params['Result'])
            # print("Received '{}'".format(recv_params))
            ws.close()
        except Exception as e:
            context = {'msg': "连接失败，请联系管理员！\n"+e}
            return render(request, 'index/error.html', context=context)

    return redirect('/status/1')


def status(request, pindex):
    solution_list = Solution.objects.all().order_by("-submitted")
    paginator = Paginator(solution_list, 10)
    if pindex == "":  # django默认返回空值，设置默认值1
        pindex = 1
    else:  # 如果有返回值，把返回值转为整数型
        int(pindex)
    page = paginator.page(pindex)  # 传递当前页的实例对象到前端
    context = {'page': page, 'type': 'list'}
    request.encoding = 'utf-8'
    return render(request, 'status/status.html', context=context)


def showsource(request, solution_id):
    context = {
        'solution_list': Solution.objects.all(),
        'solution_id': solution_id
    }
    return render(request, 'status/showsource.html', context=context)


def search(request, pindex):
    if request.method == 'POST':
        problem_id = int("0"+request.POST.get('problem_id'))
        username = request.POST.get('username')
        result = request.POST.get('result')
        lang = request.POST.get('lang')

    solution_res_obj = Solution.objects
    if problem_id != 0:
        solution_res_obj = solution_res_obj.filter(problem_id=problem_id)
    if username != "":
        solution_res_obj = solution_res_obj.filter(username=username)
    if result != "All":
        solution_res_obj = solution_res_obj.filter(result=result)
    if lang != "All":
        solution_res_obj = solution_res_obj.filter(lang=lang)
    solution_res_obj = solution_res_obj.order_by("-submitted")

    paginator = Paginator(solution_res_obj, 10)
    if pindex == "":  # django默认返回空值，设置默认值1
        pindex = 1
    else:  # 如果有返回值，把返回值转为整数型
        int(pindex)
    page = paginator.page(pindex)  # 传递当前页的实例对象到前端
    context = {
        'page': page,
        'type': 'search',
        'problem_id': problem_id,
        'username': username,
        'result': result,
        'lang': lang
    }
    request.encoding = 'utf-8'
    return render(request, 'status/status.html', context=context)
