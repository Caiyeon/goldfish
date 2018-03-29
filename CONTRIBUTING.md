### Contributing

Firstly, thank you for taking the time to improve goldfish!

First time contributors are welcomed, but please make sure your code <b>compiles</b>. This includes eliminating JSLint warnings.

Styling preferences:
  * Go code should use tabs for indentation
  * Javascript (.vue) code should use 2 spaces for indentation
  * JSLint warnings should not appear when running the frontend
  * Lines should not have trailing white space
  * No custom CSS. Rely on [Bulma CSS](https://bulma.io) for spacing, coloring, etc.

Goldfish is becoming more and more stable, and therefore places higher priority on maintainability, expressiveness, and error handling. What this means is that features need a reason for existence beyond simply "it's a feature". Features need to be <b>needed</b>. 

If you have any doubts on the necessity of a feature, please open a feature request issue first. Keep in mind many new feature requests are rejected because they fall out of scope or set a new precedent. Your feature could solve the halting problem, but still be rejected because it doesn't belong in the repository. Backend integration tests should be implemented as much as possible.
