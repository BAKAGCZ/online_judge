#include <iostream>
#include <cstdio>

using namespace std;

int n, m;//n is the number of students. m is the number of groups. 
int par[30010];
int k, g, sum;
int stu[30010];
//student 0 is recognized as a suspect in all the cases. 

int findd(int x)	//找根
{
	if (par[x] == x)	//自己是根
		return x;
	return par[x] = findd(par[x]);	//找自己的根
}

int main() 
{
	while (scanf("%d%d", &n, &m) && (n != 0 || m != 0))
	{
		for (int i = 0; i < n; i++)
			par[i] = i;

		for (int i = 0; i < m; i++)
		{
			scanf("%d", &k);

			for (int i = 0; i < k; i++)
				scanf("%d", &stu[i]);
			for (int i = 0; i < k-1; i++)
				par[findd(stu[i])] = par[findd(stu[i + 1])];
		}

		g = findd(0);
		sum = 1;
		for (int i = 1; i < n; i++)
			if (findd(i) == g)
				sum++;
		printf("%d\n", sum);
	}

	return 0;
}