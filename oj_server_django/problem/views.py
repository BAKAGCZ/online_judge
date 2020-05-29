from django.shortcuts import render
from .models import Problem
from django.core.paginator import Paginator


def problemlist(request, pindex):
    problem_list = Problem.objects.all()
    paginator = Paginator(problem_list, 10)
    if pindex == "":  # django默认返回空值，设置默认值1
        pindex = 1
    else:  # 如果有返回值，把返回值转为整数型
        int(pindex)
    page = paginator.page(pindex)  # 传递当前页的实例对象到前端
    context = {'page': page, 'type': 'list'}
    return render(request, 'problem/list.html', context=context)


def search(request, pindex):
    if request.method == 'POST':  # 0 "" All All
        search_range = request.POST['search_range']
        search_string = request.POST['search_string']
    elif request.method == 'GET':
        search_range = request.GET['search_range']
        search_string = request.GET['search_string']

    problem_res_obj = Problem.objects
    if search_string != None and search_string != "":
        if search_range == "problem_id":
            problem_res_obj = problem_res_obj.filter(id=search_string)
        if search_range == "title":
            problem_res_obj = problem_res_obj.filter(
                title__icontains=search_string)
        if search_range == "point":
            problem_res_obj = problem_res_obj.filter(point=search_string)
    else:
        problem_res_obj = problem_res_obj.all()

    paginator = Paginator(problem_res_obj, 10)
    if pindex == "":  # django默认返回空值，设置默认值1
        pindex = 1
    else:  # 如果有返回值，把返回值转为整数型
        int(pindex)
    page = paginator.page(pindex)  # 传递当前页的实例对象到前端
    context = {
        'page': page,
        'type': 'search',
        'search_range': search_range,
        'search_string': search_string
    }
    return render(request, 'problem/list.html', context=context)


def problem(request, problem_id):
    context = {
        'problem_list': Problem.objects.all(),
        'problem_id': problem_id
    }
    return render(request, 'problem/problem.html', context=context)


def submit(request, problem_id):
    context = {'problem_id': problem_id}
    return render(request, 'problem/submit.html', context=context)
