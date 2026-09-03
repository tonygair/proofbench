--  <vc-preamble>
package Np_Shape_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   --  Bound on each dimension of the nested-array variant, kept small so the
   --  three-dimensional alternative of Arrays stays a finite object.
   Max_Nested : constant := 32;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Dim_Type is Natural range 0 .. Max_Index + 1;
   subtype Nested_Dim is Natural range 0 .. Max_Nested;

   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;

   function Matrix_Size (M : Matrix) return Natural is
     (M'Length (1) * M'Length (2));

   --  A shape is Dafny's seq<nat>, indexed from zero.
   type Shape_Array is array (Natural range <>) of Dim_Type;

   type Array_1D is array (Positive range <>) of Value_Type;
   type Array_2D is array (Positive range <>, Positive range <>) of Value_Type;
   type Array_3D is
     array (Positive range <>, Positive range <>, Positive range <>)
       of Value_Type;

   type Ndim_Kind is (One, Two, Three);

   --  Dafny's Arrays datatype: a one-, two- or three-dimensional array.
   --  Dafny's nested sequences may be ragged while SPARK arrays are
   --  rectangular, but the specification below only ever inspects arr[0] and
   --  arr[0][0], so the rectangular model states exactly the same thing.
   type Arrays (Kind : Ndim_Kind; D1, D2, D3 : Nested_Dim) is record
      case Kind is
         when One =>
            A1 : Array_1D (1 .. D1);
         when Two =>
            A2 : Array_2D (1 .. D1, 1 .. D2);
         when Three =>
            A3 : Array_3D (1 .. D1, 1 .. D2, 1 .. D3);
      end case;
   end record;

   function Arrays_Ndim (A : Arrays) return Positive is
     (case A.Kind is
         when One   => 1,
         when Two   => 2,
         when Three => 3);
--  </vc-preamble>

--  <vc-spec>
   --  Dafny returns a seq whose length varies with the argument, so the
   --  SPARK analogue is a function returning an unconstrained array.
   function Shape_Arrays (A : Arrays) return Shape_Array with
     Post => Shape_Arrays'Result'First = 0
             and then Shape_Arrays'Result'Length = Arrays_Ndim (A)
             and then (if A.Kind = One
                       then Shape_Arrays'Result (0) = A.A1'Length)
             and then (if A.Kind = Two
                       then Shape_Arrays'Result (0) = A.A2'Length (1)
                            and then Shape_Arrays'Result (1) =
                              (if A.A2'Length (1) > 0
                               then A.A2'Length (2) else 0))
             and then (if A.Kind = Three
                       then Shape_Arrays'Result (0) = A.A3'Length (1)
                            and then Shape_Arrays'Result (1) =
                              (if A.A3'Length (1) > 0
                               then A.A3'Length (2) else 0)
                            and then Shape_Arrays'Result (2) =
                              (if A.A3'Length (1) > 0
                                 and then A.A3'Length (2) > 0
                               then A.A3'Length (3) else 0));

   function Shape_Matrix (A : Matrix) return Shape_Array with
     Post => Shape_Matrix'Result'First = 0
             and then Shape_Matrix'Result'Length = 2
             and then Shape_Matrix'Result (0) = A'Length (1)
             and then Shape_Matrix'Result (1) = A'Length (2);

end Np_Shape_Spec;

package body Np_Shape_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Shape_Arrays (A : Arrays) return Shape_Array is
   begin
      pragma Assume (False);
      return (0 => 0);
   end Shape_Arrays;

      function Shape_Matrix (A : Matrix) return Shape_Array is
   begin
      pragma Assume (False);
      return (0 => 0, 1 => 0);
   end Shape_Matrix;
--  </vc-code>

--  <vc-postamble>
end Np_Shape_Spec;
--  </vc-postamble>
