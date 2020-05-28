from django.urls import path
from . import views

urlpatterns = [
    path('list/<int:pindex>/', views.problemlist, name='problemlist'),
    path('<int:problem_id>/', views.problem, name='problem'),
    path('search/<int:pindex>/', views.search, name='search'),
    path('submit/<int:problem_id>/', views.submit, name='submit'),
]
