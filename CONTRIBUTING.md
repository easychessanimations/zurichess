Contributions to zurichess are welcomed. These are some guidelines to follow:

* Keep the patch as simple as possible. Get familiar with this article
http://tirania.org/blog/archive/2010/Dec-31.html.
* go fmt and go test your code. Add tests if possible.
* The code will be reviewed. Complicated patches will be rejected, but
feedback will be given. Changes that break encapsulation boundaries (e.g.
adding evaluation elements to the board logic) will be rejected.
* Evaluation function can be trained automatically using the
[tuner](https://bitbucket.org/zurichess/tuner). See engine/material.go
for a description of the evaluation function.
* Search parameters are tuned manually.
* Improvement patches are tested at two different time controls using
SPRT stopping rule. To reduce the cluster testing time please include in
the pull request the results of a match of at least 5000 games at 40/5+0.05.
* Regression patches are tested at short time control only.

Things that can be improved:

* King safety.
* Passed pawns evaluation, especially in the end game.
* Material imbalance.
* LMR, NMP and FP conditions.
* SEE prunings in QS to handle some tactics such as discovered attacks.
* End game evaluation.


[zuritest](https://bitbucket.org/zurichess/zuritest) is the cluster framework
used to test zurichess. Due to some missing features there is no public instance
of the framework.

