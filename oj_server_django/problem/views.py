from django.shortcuts import render
from .models import Problem
# from django.core.paginator import Paginator

# def problemlist(request):
#     problem_list = Problem.objects.all()
#     paginator = Paginator(problem_list, 5)
#     if pindex == "":  # django默认返回空值，设置默认值1
#         pindex = 1
#     else:  # 如果有返回值，把返回值转为整数型
#         int(pindex)
#     page = paginator.page(pindex)  # 传递当前页的实例对象到前端
#     context = {'page': page}
#     return render(request, 'problem/list.html', context=context)


def problemlist(request):
    context = {'problem_list': Problem.objects.all()}
    return render(request, 'problem/list.html', context=context)


def search(request):
    if request.method == 'POST':  # 0 "" All All
        search_range = request.POST['search_range']
        search_string = request.POST['search_string']

    problemResObj = Problem.objects
    if search_string != None and search_string != "":
        if search_range == "problem_id":
            problemResObj = problemResObj.filter(id=search_string)
        if search_range == "title":
            problemResObj = problemResObj.filter(
                title__icontains=search_string)
        if search_range == "point":
            problemResObj = problemResObj.filter(point=search_string)
        context = {'problem_list': problemResObj,
                   'search_range': search_range, 'search_string': search_string}
    else:
        context = {'problem_list': problemResObj.all(),
                   'search_range': search_range, 'search_string': search_string}
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
